package identity

import (
	"net/http"
)

func RegisterHandlers(mux *http.ServeMux, uuid string) {
	mux.HandleFunc("/device.xml", func(w http.ResponseWriter, _ *http.Request) {
		PolishHeaders(w)
		w.Header().Set("Content-Type", "text/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(DeviceXML(uuid)))
	})

	mux.HandleFunc("/cd/scpd.xml", func(w http.ResponseWriter, _ *http.Request) {
		PolishHeaders(w)
		w.Header().Set("Content-Type", "text/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(EmptySCPD))
	})

	mux.HandleFunc("/cm/scpd.xml", func(w http.ResponseWriter, _ *http.Request) {
		PolishHeaders(w)
		w.Header().Set("Content-Type", "text/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(EmptySCPD))
	})
}
