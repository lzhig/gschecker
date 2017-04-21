package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"globaltedinc/framework/network"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os/exec"
	"time"

	"github.com/axgle/mahonia"
)

func execCommand(commandName string, params []string) (output string, err error) {
	cmd := exec.Command(commandName, params...)

	//显示运行的命令
	//fmt.Println(cmd.Args)

	stdout, err := cmd.StdoutPipe()

	if err != nil {
		return "", err
	}

	if err = cmd.Start(); err != nil {
		return "", err
	}

	reader := bufio.NewReader(stdout)

	//实时循环读取输出流中的一行内容
	enc := mahonia.NewDecoder("gbk")
	for {
		line, err2 := reader.ReadString('\n')
		if err2 != nil || io.EOF == err2 {
			break
		}
		output += enc.ConvertString(line)
	}

	if err = cmd.Wait(); err != nil {
		return "", err
	}

	return output, err
}

func getSystemInfo() (string, error) {
	if r, err := execCommand("cmd", []string{"/C", "systeminfo"}); err != nil {
		return "", err
	} else {
		return r, nil
	}
}

func getSystemInfoToReport(r *reportT) error {
	ret, err := getSystemInfo()
	if err != nil {
		r.SystemInfo = fmt.Sprint(err)
		return err
	}

	r.SystemInfo = ret
	return nil
}

func getExternalIP() (string, error) {
	resp, err := http.Get("http://myexternalip.com/json")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	r, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	//fmt.Println(r)

	var ip map[string]string
	err = json.Unmarshal(r, &ip)
	if err != nil {
		return "", err
	}

	ret, ok := ip["ip"]
	if ok {
		return ret, nil
	}
	return "", nil

}

func getExternalIPToReport(r *reportT) error {
	ret, err := getExternalIP()
	if err != nil {
		r.PublicIP = fmt.Sprint(err)
		return err
	}

	r.PublicIP = ret

	fmt.Println("IP:", ret)
	return nil
}

func getDNSServer() (s string, err error) {
	r, err := execCommand("cmd", []string{"/C", "nslookup www.gs108.com"})
	if err != nil {
		return "", err
	}
	return r, nil

}

func getDNSServerToReport(r *reportT) error {
	ret, err := getDNSServer()
	if err != nil {
		r.UserDNSServer = fmt.Sprint(err)
		return err
	}

	r.UserDNSServer = ret
	return nil
}

func resolveDNS(dns string) (string, error) {
	r, err := net.LookupHost("www.gs108.com")
	if err != nil || len(r) == 0 {
		return "", err
	}
	return fmt.Sprint(r[0]), nil

}

func resolveDNSGS108() (string, error) {
	return resolveDNS("www.gs108.com")
}

func resolveDNSGS108ToReport(r *reportT) error {
	t := time.Now()
	ret, err := resolveDNSGS108()
	if err != nil {
		r.Gs108DNS.Error = fmt.Sprint(err)
		return err
	}
	r.Gs108DNS.IP = ret
	r.Gs108DNS.Time = fmt.Sprintf("%dms", time.Now().Sub(t)/time.Millisecond)
	return nil
}

func resolveDNSSpecific() (s string, err error) {
	// 从服务器获取需要解析的DNS列表
	lists := []string{"testd.ttl01.com", "play.gs108.com", "testd.ttl01.com"}
	for ndx := range lists {
		l := lists[ndx]
		s += l + ":"
		t := time.Now()
		if r, err := net.LookupHost(l); err != nil {
			s += fmt.Sprint(err)
		} else {
			delta := time.Now().Sub(t)
			s += fmt.Sprint(r)
			s += fmt.Sprintf(" time:%s\n", convertTime(delta, 2))
		}
	}

	return s, nil
}

func resolveDNSSpecificToReport(c configT, r *reportT) error {
	r.ResolveDNSList = make(map[string]DNSReportT, len(c.ResolveDNSList))
	for ndx := range c.ResolveDNSList {
		url := c.ResolveDNSList[ndx]
		t := time.Now()
		ret, err := resolveDNS(url)
		var dns DNSReportT
		if err != nil {
			dns.Error = fmt.Sprint(err)
		} else {
			dns.IP = ret
			dns.Time = fmt.Sprintf("%dms", time.Now().Sub(t)/time.Millisecond)
		}

		r.ResolveDNSList[url] = dns
	}

	return nil
}

func resolveRoutersToReport(c configT, r *reportT) error {
	r.IPRouterList = make(map[string]string, len(c.IPRouterList))
	for ndx := range c.IPRouterList {
		addr := c.IPRouterList[ndx]
		ret, err := execCommand("cmd", []string{"/C", "WinMTRCmd.exe -c 1 -r", addr})
		if err != nil {
			r.IPRouterList[addr] = fmt.Sprint(err)
		} else {
			r.IPRouterList[addr] = ret
		}
	}
	return nil
}

type timeFormatDef struct {
	level  time.Duration
	symbol string
}

var timeFormatDefines = []timeFormatDef{
	{time.Hour, "h"},
	{time.Minute, "m"},
	{time.Second, "s"},
	{time.Millisecond, "ms"},
	{time.Microsecond, "µs"},
	{time.Nanosecond, "ns"},
}

func convertTime(t time.Duration, level uint32) (s string) {
	if t == 0 {
		return "0ns"
	}

	var c uint32
	for ndx := range timeFormatDefines {
		d := timeFormatDefines[ndx]
		if t > d.level {
			a := t / d.level
			s += fmt.Sprintf("%d%s", a, d.symbol)
			t -= a * d.level
			c++
		}

		if level > 0 && level <= c {
			return s
		}
	}

	return s
}

func checkDownloadSpeed() (s string, err error) {
	lists := []string{
		"http://7xsx8e.com2.z0.glb.qiniucdn.com/media/flash/game/1234567890abcdef/TexasPokerRoom.swf?v=2122",
		"http://7xsx8e.com2.z0.glb.qiniucdn.com/media/flash/game/1234567890abcdef/LoadingFileAction.swf?v=2122",
	}

	for ndx := range lists {
		l := lists[ndx]

		s += "url: " + l + "\n"

		t := time.Now()

		res, err := http.Get(l)
		if err != nil {
			return "", err
		}
		defer res.Body.Close()

		d, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return "", err
		}

		delta := time.Now().Sub(t)
		k := float64(len(d)) / 1024.0
		t1 := float64(delta) / float64(time.Second)
		speed := k / t1
		s += fmt.Sprintf("file size: %d, time:%s, speed: %fK/s\n", len(d), convertTime(delta, 2), speed)
	}
	return s, nil
}

func checkDownloadSpeedToReport(c configT, r *reportT) error {
	r.CheckSpeedURLList = make(map[string]URLSpeedResultItemT, len(c.CheckSpeedURLList))
	for ndx := range c.CheckSpeedURLList {
		url := c.CheckSpeedURLList[ndx]
		var result URLSpeedResultItemT
		defer func(l map[string]URLSpeedResultItemT, url string, result *URLSpeedResultItemT) {
			l[url] = *result
		}(r.CheckSpeedURLList, url, &result)

		t := time.Now()
		res, err := http.Get(url)
		if err != nil {
			result.Error = fmt.Sprint(err)
			continue
		}
		defer res.Body.Close()

		d, err := ioutil.ReadAll(res.Body)
		if err != nil {
			result.Error = fmt.Sprint(err)
			continue
		}

		delta := time.Now().Sub(t)
		var speed uint32
		if delta == 0 {
			speed = 0
		} else {
			speed = uint32((float64(len(d)) / 1024.0) / (float64(delta) / float64(time.Second)))
		}

		result.Size = uint32(len(d))
		result.Time = uint32(delta / time.Millisecond)
		result.Speed = speed
	}

	return nil
}

func checkGameServerSpeed(addr string, size uint32, count uint32, timeout uint32) (result GameServerSpeedResultItemT) {
	var client network.TCPClient

	c := make(chan GameServerSpeedResultItemT, 1)
	quit := make(chan bool, 1)

	sendCount := 0
	t1 := time.Now()
	maxTime := uint32(0)
	minTime := uint32(999999)
	totalTime := 0

	if err := client.Connect(addr, timeout,
		func(err error) {
			if result.RecvCount < count {
				result.Error = fmt.Sprint(err)
				result.MinTime = minTime
				result.MaxTime = maxTime
				if result.RecvCount == 0 {
					result.AveTime = 0
				} else {
					result.AveTime = uint32(totalTime / int(result.RecvCount))
				}
				fmt.Println("error:", result.Error)
				c <- result
			}
			quit <- true
		},

		func(packet *network.Packet) {
			t := int(time.Now().Sub(t1) / time.Millisecond)
			if t == 0 {
				t = 1
			}
			totalTime += t
			if t > int(maxTime) {
				maxTime = uint32(t)
			}

			if t < int(minTime) {
				minTime = uint32(t)
			}

			result.RecvCount++
			//fmt.Println(result.RecvCount)

			if result.RecvCount >= count {
				result.MinTime = minTime
				result.MaxTime = maxTime
				result.AveTime = uint32(totalTime / int(result.RecvCount))
				c <- result
				return
			}
			t1 = time.Now()
			client.SendPacket(packet)
		}); err != nil {
		result.Error = fmt.Sprint(err)
		return result
	}

	b := make([]byte, size)
	p := network.Packet{}
	p.Attach(b)
	t1 = time.Now()
	client.SendPacket(&p)
	sendCount++

	select {
	case result = <-c:
		//fmt.Println(result)

	case <-time.After(time.Duration(count*500) * time.Millisecond):
		fmt.Println("Timeout.")
	}

	client.Disconnect()
	//fmt.Println("connection closing.")
	<-quit
	//fmt.Println("connection closed.")

	return result
}

func checkGameServerSpeedToReport(c configT, r *reportT) error {
	r.GameServerList = make(map[string]GameServerSpeedResultItemT, len(c.GameServerList))
	for k, v := range c.GameServerList {
		fmt.Println("检测", k, "...")
		r.GameServerList[k] = checkGameServerSpeed(k, v.Size, v.Count, v.Timeout)
	}

	return nil
}
