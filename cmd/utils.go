package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/grasparv/go-chromecast/application"
	castdns "github.com/grasparv/go-chromecast/dns"
	"github.com/grasparv/go-chromecast/storage"
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

var (
	cache = storage.NewStorage()
)

type CachedDNSEntry struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
	Addr string `json:"addr"`
	Port int    `json:"port"`
}

func (e CachedDNSEntry) GetUUID() string {
	return e.UUID
}

func (e CachedDNSEntry) GetName() string {
	return e.Name
}

func (e CachedDNSEntry) GetAddr() string {
	return e.Addr
}

func (e CachedDNSEntry) GetPort() int {
	return e.Port
}

func castApplication(cmd *cobra.Command, args []string) (*application.Application, error) {
	deviceName, _ := cmd.Flags().GetString("device-name")
	deviceUuid, _ := cmd.Flags().GetString("uuid")
	device, _ := cmd.Flags().GetString("device")
	debug, _ := cmd.Flags().GetBool("debug")
	disableCache, _ := cmd.Flags().GetBool("disable-cache")
	addr, _ := cmd.Flags().GetString("addr")
	port, _ := cmd.Flags().GetString("port")
	iface, _ := cmd.Flags().GetString("iface")

	var entry castdns.CastDNSEntry
	// If no address was specified, attempt to determine the address of any
	// local chromecast devices.
	if addr == "" {
		// If a device name or uuid was specified, check the cache for the ip+port
		found := false
		if !disableCache && (deviceName != "" || deviceUuid != "") {
			entry = findCachedCastDNS(deviceName, deviceUuid)
			found = entry.GetAddr() != ""
		}
		if !found {
			var err error
			if entry, err = findCastDNS(device, deviceName, deviceUuid); err != nil {
				return nil, errors.Wrap(err, "unable to find cast dns entry")
			}
		}
		if !disableCache {
			cachedEntry := CachedDNSEntry{
				UUID: entry.GetUUID(),
				Name: entry.GetName(),
				Addr: entry.GetAddr(),
				Port: entry.GetPort(),
			}
			cachedEntryJson, _ := json.Marshal(cachedEntry)
			cache.Save(getCacheKey(cachedEntry.UUID), cachedEntryJson)
			cache.Save(getCacheKey(cachedEntry.Name), cachedEntryJson)
		}
		if debug {
			fmt.Printf("using device name=%s addr=%s port=%d uuid=%s\n", entry.GetName(), entry.GetAddr(), entry.GetPort(), entry.GetUUID())
		}
	} else {
		p, err := strconv.Atoi(port)
		if err != nil {
			return nil, errors.Wrap(err, "port needs to be a number")
		}
		entry = CachedDNSEntry{
			Addr: addr,
			Port: p,
		}
	}
	app := application.NewApplication(iface, debug, disableCache)
	if err := app.Start(entry); err != nil {
		// NOTE: currently we delete the dns cache every time we get
		// an error, this is to make sure that if the device gets a new
		// ipaddress we will invalidate the cache.
		cache.Save(getCacheKey(entry.GetUUID()), []byte{})
		cache.Save(getCacheKey(entry.GetName()), []byte{})
		return nil, err
	}
	return app, nil
}

func getCacheKey(suffix string) string {
	return fmt.Sprintf("cmd/utils/dns/%s", suffix)
}

func findCachedCastDNS(deviceName, deviceUuid string) castdns.CastDNSEntry {
	for _, s := range []string{deviceName, deviceUuid} {
		cacheKey := getCacheKey(s)
		if b, err := cache.Load(cacheKey); err == nil {
			cachedEntry := CachedDNSEntry{}
			if err := json.Unmarshal(b, &cachedEntry); err == nil {
				return cachedEntry
			}
		}
	}
	return CachedDNSEntry{}
}

func findCastDNS(device, deviceName, deviceUuid string) (castdns.CastDNSEntry, error) {
	dnsEntries := castdns.FindCastDNSEntries()
	switch l := len(dnsEntries); l {
	case 0:
		return castdns.CastEntry{}, errors.New("no cast dns entries found")
	default:
		for _, d := range dnsEntries {
			if (deviceUuid != "" && d.UUID == deviceUuid) || (deviceName != "" && d.DeviceName == deviceName) || (device != "" && d.Device == device) {
				return d, nil
			}
		}

		if l == 1 {
			return dnsEntries[0], nil
		}

		fmt.Printf("Found %d cast dns entries, select one:\n", l)
		for i, d := range dnsEntries {
			fmt.Printf("%d) device=%q device_name=%q address=\"%s:%d\" status=%q uuid=%q\n", i+1, d.Device, d.DeviceName, d.AddrV4, d.Port, d.Status, d.UUID)
		}
		reader := bufio.NewReader(os.Stdin)
		for {
			fmt.Printf("Enter selection: ")
			text, err := reader.ReadString('\n')
			if err != nil {
				fmt.Printf("error reading console: %v\n", err)
				continue
			}
			i, err := strconv.Atoi(strings.TrimSpace(text))
			if err != nil {
				continue
			} else if i < 1 || i > l {
				continue
			}
			return dnsEntries[i-1], nil
		}
	}
	return castdns.CastEntry{}, errors.New("no cast dns entries found")
}
