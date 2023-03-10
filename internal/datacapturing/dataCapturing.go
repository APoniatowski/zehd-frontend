package datacapturing

import (
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	timeSmithy "github.com/aws/smithy-go/time"

	"poniatowski-dev/internal/backendconnector"
	"poniatowski-dev/internal/logging"
)

// CollectData - Collects all data from http request, parses it and sends it to the database
func CollectData(r *http.Request) {
	var rD backendconnector.RequestData
	var errHostname error
	var waitGroup sync.WaitGroup

	// get all data first
	rD.Path = r.URL.Path
	rD.Method = r.Method
	rD.XForwardFor = r.Header.Get("X-Forwarded-For")
	rD.XRealIP = r.Header.Get("X-Real-IP")
	rD.Via = r.Header.Get("Via")
	rD.UserAgent = r.Header.Get("User-Agent")
	rD.Age = r.Header.Get("Age")
	rD.CFIPCountry = r.Header.Get("CF-IPCountry")
	var tmpPort string
	rD.IP, tmpPort, _ = net.SplitHostPort(r.RemoteAddr)
	rD.Port, _ = strconv.Atoi(tmpPort)
	headerTime := r.Header.Get("Date")
	timeNow, timeErr := timeSmithy.ParseHTTPDate(headerTime)
	if timeErr != nil {
		logging.LogIt("CollectData", "WARNING", "unable to parse header time and date")
		timeNow = time.Now()
	}
	rD.TimeDate = timeNow.UnixNano()

	rD.FrontendName, errHostname = os.Hostname()
	if errHostname != nil {
		logging.LogIt("CollectData", "ERROR", "unable to get frontend hostname")
	}
	// do checks on data obtained
	netIP := net.ParseIP(rD.IP)
	if netIP == nil {
		logging.LogIt("CollectData", "WARNING", "remoteAddr ["+rD.IP+"] : no valid ip found")
	}
	splitIPs := strings.Split(rD.XForwardFor, ",")
	for _, ip := range splitIPs {
		netIPs := net.ParseIP(ip)
		if netIPs == nil {
			logging.LogIt("CollectData", "WARNING", "x-forward-for ["+ip+"] : no valid ip found")
		}
	}
	// send data to database
	go rD.DBConnector(&waitGroup)
	waitGroup.Wait()
}
