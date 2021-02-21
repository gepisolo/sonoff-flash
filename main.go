package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sync/semaphore"
)

const (
	StatusMainMenu                 = "main-menu"
	StatusScan4ITEAD               = "scan-itead"
	StatusSelectYourNetwork        = "select-network"
	StatusFirstFindDevice          = "first-find-device"
	StatusNetworkInterfaceSelected = "net-interface-selected"
	StatusExit                     = "exit"
)

var status = StatusMainMenu

func main() {
	for {
		fmt.Println("**** eval")
		switch status {
		case StatusExit:
			os.Exit(0)
		case StatusMainMenu:
			MainMenu()
		case StatusFirstFindDevice:
			FindNewDevice()
		case StatusScan4ITEAD:
			FindITEADNetwork()
			os.Exit(0)
		}
	}

}

func FindITEADNetwork() {
	fmt.Println("Scanning WIFI for 30 seconds")
}

func FindNewDevice() {
	fmt.Println("Please turn on your device.")
	fmt.Println("Long press its button (>5sec).")
	fmt.Println("It should blink rapidly and continuously (if not, long press again).")

	menu := []*Menu{
		&Menu{
			Idx:                  1,
			Label:                "When device is ready, press (1)",
			Obj:                  nil,
			LeaveBlankLineBefore: false,
			LeaveBlankLineAfter:  false,
		},
		&Menu{
			Idx:                  0,
			Label:                "Exit",
			Obj:                  nil,
			LeaveBlankLineBefore: true,
			LeaveBlankLineAfter:  true,
		},
	}
	choice := GetChoiceFromMenu("FOLLOW INSTRUCTIONS", menu)
	switch choice.Idx {
	case 1:
		status = StatusScan4ITEAD
		return
	}
	status = StatusExit
}

func MainMenu() {
	menu := []*Menu{
		&Menu{
			Idx:                  1,
			Label:                "First connect device to your network",
			Obj:                  nil,
			LeaveBlankLineBefore: false,
			LeaveBlankLineAfter:  false,
		},
		&Menu{
			Idx:                  2,
			Label:                "Find device and unlock OTA firmware flash",
			Obj:                  nil,
			LeaveBlankLineBefore: false,
			LeaveBlankLineAfter:  false,
		},
		&Menu{
			Idx:                  3,
			Label:                "Flash device (firmware needed)",
			Obj:                  nil,
			LeaveBlankLineBefore: false,
			LeaveBlankLineAfter:  false,
		},
		&Menu{
			Idx:                  0,
			Label:                "Exit",
			Obj:                  nil,
			LeaveBlankLineBefore: true,
			LeaveBlankLineAfter:  true,
		},
	}
	choice := GetChoiceFromMenu("MAIN MENU:", menu)
	switch choice.Idx {
	case 1:
		status = StatusFirstFindDevice
		return
	}
	status = StatusExit
}

func showScan(ip *addr) {
	ps := &PortScanner{
		ip:   ip.pretty,
		port: 8081,
	}
	found := ps.Start(500 * time.Millisecond)
	if len(found) == 0 {
		fmt.Println("No device found")
		return
	}

	for _, f := range found {
		fmt.Println("FOUND ", f)
	}

}

type Menu struct {
	Idx                  uint
	Label                string
	Obj                  interface{}
	LeaveBlankLineBefore bool
	LeaveBlankLineAfter  bool
}

func GetChoiceFromMenu(title string, menu []*Menu) *Menu {
	fmt.Println(title)
	buf := bufio.NewReader(os.Stdin)
	for _, m := range menu {
		if m.LeaveBlankLineBefore {
			fmt.Println()
		}
		fmt.Println()
		fmt.Printf("%d) %s", m.Idx, m.Label)
		if m.LeaveBlankLineAfter {
			fmt.Println()
		}
	}
	fmt.Print("Make your choice > ")

	sentence, err := buf.ReadString('\n')
	if err != nil {
		panic(err)
	}

	sentence = strings.TrimRightFunc(sentence, func(c rune) bool {
		//In windows newline is \r\n
		return c == '\r' || c == '\n'
	})

	num, err := strconv.Atoi(sentence)
	if err != nil {
		fmt.Println("Cannot understand your choice (only numbers please)")
		return GetChoiceFromMenu(title, menu)
	}

	idx := uint(num)
	for _, m := range menu {
		if m.Idx == idx {
			return m
		}
	}

	fmt.Println("Cannot understand your choice")
	return GetChoiceFromMenu(title, menu)

}

func printNetInterfacesMenu() *addr {
	buf := bufio.NewReader(os.Stdin)
	ads := localAddresses()
	if len(ads) == 0 {
		fmt.Println("NO Network interfaces")
		return nil
	}

	fmt.Println("Choose your network interface:")
	printAddresses(ads)

	fmt.Print("\nYour choice: > ")
	sentence, err := buf.ReadString('\n')
	if err != nil {
		fmt.Println(err)
		return nil
	}

	sentence = strings.TrimRightFunc(sentence, func(c rune) bool {
		//In windows newline is \r\n
		return c == '\r' || c == '\n'
	})

	num, err := strconv.Atoi(sentence)
	if err != nil {
		fmt.Println("Cannot understand your choice (only numbers please)")
		fmt.Println(err)
		return nil
	}

	idx := uint(num)
	ip, ok := ads[idx]
	if !ok {
		fmt.Println("Out of range choice")
	}
	return ip

}

func printAddresses(addresses map[uint]*addr) {
	l := len(addresses)
	for n := uint(0); n < uint(l); n++ {
		idx := n + 1
		if a, ok := addresses[idx]; ok {
			fmt.Printf("%d) %s:  %s \n", idx, a.name, a.pretty)
		}
	}
}

func localAddresses() map[uint]*addr {
	interfaces, err := net.Interfaces()
	ret := make(map[uint]*addr, 0)
	if err != nil {
		fmt.Print(fmt.Errorf("localAddresses: %+v\n", err.Error()))
		return nil
	}
	n := uint(0)
	for _, i := range interfaces {
		addresses, err := i.Addrs()
		if err != nil {
			fmt.Print(fmt.Errorf("localAddresses: %+v\n", err.Error()))
			continue
		}
		for _, a := range addresses {
			switch v := a.(type) {
			/*case *net.IPAddr:
			fmt.Printf("A) %v : %s (%s)\n", i.Name, v, v.IP.DefaultMask())*/

			case *net.IPNet:
				if v.IP.To4() == nil {
					continue
				}
				n++
				ret[n] = &addr{
					IPNet: net.IPNet{
						IP:   v.IP,
						Mask: v.Mask,
					},
					pretty: fmt.Sprintf("%s", v),
					name:   i.Name,
				}
				//fmt.Printf("N) %v : %s [%v/%v]\n", i.Name, v, v.IP, v.Mask)
			}

		}
	}

	return ret
}

type addr struct {
	net.IPNet
	name   string
	pretty string
}

/* PORT SCANNER */

type PortScanner struct {
	ip   string
	port int
	lock *semaphore.Weighted
}

func ScanPort(ip string, port int, timeout time.Duration) bool {
	target := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.DialTimeout("tcp", target, timeout)
	fmt.Println("Scanning ... ", ip, " on port ", port)
	if err != nil {
		if strings.Contains(err.Error(), "too many open files") {
			time.Sleep(timeout)
			return ScanPort(ip, port, timeout)
		} else {
			return false
			//fmt.Println(port, "closed")
		}
	}

	conn.Close()
	return true
}

func (ps *PortScanner) Start(timeout time.Duration) map[int]string {
	hosts, num, err := Hosts(ps.ip)
	if err != nil {
		fmt.Println("Error: cannot find hosts from your network")
		return nil
	}
	if num > 255 {
		fmt.Println("Error: too many ip addresses: ", num)
		return nil
	}

	n := 1
	ret := make(map[int]string, 0)
	for _, ip := range hosts {
		if ScanPort(ip, ps.port, timeout) {
			ret[n] = ip
			n++
		}
	}

	return ret
}

func Hosts(cidr string) ([]string, int, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, 0, err
	}

	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}

	// remove network address and broadcast address
	lenIPs := len(ips)
	switch {
	case lenIPs < 2:
		return ips, lenIPs, nil

	default:
		return ips[1 : len(ips)-1], lenIPs - 2, nil
	}
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

/*
POST
http://10.10.7.1/ap_diy
{"ssid":"Gepisolo-VS","password":"viasilvani33#"}
fetch("http://10.10.7.1/ap_diy",{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({ssid:e,password:t})})
*/
