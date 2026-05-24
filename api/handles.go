package api

import (
	"net/http"

	"gokongres/api/arrivals"
	"gokongres/api/auth"
	"gokongres/api/buffer"
	"gokongres/api/config"
	"gokongres/api/ia"
	"gokongres/api/limits"
	"gokongres/api/monitoring"
	"gokongres/api/pk"
	"gokongres/api/rja"
	"gokongres/api/sector"
	"gokongres/api/sra"
	"gokongres/api/srp"
	"gokongres/api/terminals"
)

// RegisterHandlers rejestruje endpointy HTTP używane przez serwera.
func RegisterHandlers(host string, port int) {
	r := http.NewServeMux()

	// Dodaj session middleware
	// r.Use(sessions.SessionMiddleware)

	// config endpoints
	r.HandleFunc("/config/all", config.Get_AllConfig) // zamiast /config/tury
	r.HandleFunc("/config/active/tura", config.Get_ActiveTura)

	// Auth endpoints
	r.HandleFunc("/auth/login", auth.LoginHandler)
	r.HandleFunc("/auth/admin", auth.AdminHandler)
	r.HandleFunc("/auth/logout", auth.LogoutHandler)
	r.HandleFunc("/auth/permissions", auth.PermissionsHandler)

	// arrivals endpoints (100%)
	r.HandleFunc("/arrivals/all", arrivals.Get_All) // GET
	r.HandleFunc("/arrivals/set", arrivals.Set)     // POST

	// limits endpoints
	r.HandleFunc("/limits/zbory", limits.Get_Zbory)
	r.HandleFunc("/limits/zbory/setlimit", limits.Post_SetZboryLimit)
	r.HandleFunc("/limits/dzialy", limits.Get_Dzialy)
	r.HandleFunc("/limits/dzialy/setlimit", limits.Post_SetDzialyLimit)

	// PK (parking księżycowy - działy) endpoints (10 z 11)
	r.HandleFunc("/pk/hints", pk.Get_Hints)
	r.HandleFunc("/pk/login", pk.Get_Login)
	r.HandleFunc("/pk/all", pk.Get_LoadAll)
	r.HandleFunc("/pk/create", pk.Get_CreatePassID)
	r.HandleFunc("/pk/find", pk.Get_FindPassID)
	r.HandleFunc("/pk/read/{pk_id}", pk.Get_ReadPassData)
	r.HandleFunc("/pk/update", pk.Post_UpdatePassData)
	r.HandleFunc("/pk/delete/{pk_id}", pk.Get_DeletePass)
	r.HandleFunc("/pk/download/{pk_id}", pk.Get_DownloadPassData)
	r.HandleFunc("/pk/isfreepass/{dep_name}/{tura}", pk.Get_IsFreePass)
	r.HandleFunc("/pk/check/{pass_nr}/{regnum1}/{regnum2}/{regnum3}", pk.Get_CheckPass)

	// monitoring endpoints (1 z 2)
	r.HandleFunc("/monitoring/states", monitoring.Get_StatesRepo)
	r.HandleFunc("/monitoring/terminals", monitoring.Get_TerminalsList)

	// buffer endpoints (2 z 7)
	r.HandleFunc("/buffer/all", buffer.Get_AllShortList)                // GET
	r.HandleFunc("/buffer/fullinfo/{terminal_id}", buffer.Get_FullInfo) // GET
	r.HandleFunc("/buffer/states/{terminal_id}", buffer.Get_States)     // GET
	// r.HandleFunc("/buffer/notify/nobus/{rja_id}", buffer.) // GET
	// r.HandleFunc("/buffer/notify/inbuffer/{rja_id}", buffer.) // GET
	// r.HandleFunc("/buffer/notify/secondcircle/{rja_id}", buffer.) // GET
	// r.HandleFunc("/buffer/notify/sendtosector/{rja_id}", buffer.) // GET

	// terminals endpoints
	r.HandleFunc("/terminals/all", terminals.Get_AllList)
	r.HandleFunc("/terminals/fullinfo/{terminal_id}", terminals.Get_FullInfo)

	// ia endpoints (0 z 2)
	r.HandleFunc("/ia/list/{congregation_name}", ia.Get_List)
	r.HandleFunc("/ia/download/{sra_id}", ia.Get_Download)

	// rja endpoints (7 z 9)
	r.HandleFunc("/rja/zbory", rja.Get_CongregationList)
	r.HandleFunc("/rja/sra/{tura_id}", rja.Get_SraList)
	r.HandleFunc("/rja/terminals", rja.Get_TerminalsList)
	r.HandleFunc("/rja/sectors/{terminal_id}", rja.Get_SectorsList)
	r.HandleFunc("/rja/buses/{sid}/{tura_id}", rja.Get_BusesOfSector)
	r.HandleFunc("/rja/buses/used/{tura_id}", rja.Get_BusesUsed)
	r.HandleFunc("/rja/buses/save", rja.Get_BusesSave)

	// sra endpoints (3 z 7)
	r.HandleFunc("/sra/search/congregations/{pattern}", sra.Get_SearchCongregationsByPattern)
	r.HandleFunc("/sra/submit/bus", sra.Post_SubmitBus)
	r.HandleFunc("PUT /sra/submit/nobus/{congregation_name}", sra.Put_SubmitNoBus)
	r.HandleFunc("POST /sra/check_pilot_duplicate", sra.Post_IsPilotDuplicate)
	r.HandleFunc("/sra/table", sra.Get_Table)
	r.HandleFunc("/sra/save", sra.Post_Save)

	// srp endpoints (4 z 12)
	r.HandleFunc("/srp/zbory", srp.Get_CongregationList)
	r.HandleFunc("/srp/all", srp.Get_AllList)
	r.HandleFunc("/srp/create", srp.Post_Create)
	r.HandleFunc("/srp/find", srp.Post_FindPassID)
	r.HandleFunc("/srp/delete/{srp_id}", srp.Get_Delete)
	r.HandleFunc("/srp/isfreepass/{congregation_name}", srp.Get_IsFreePass)
	r.HandleFunc("/srp/limit/{congregation_name}", srp.Get_UsingLimit)
	r.HandleFunc("/srp/limit/change", srp.Post_RequestNewLimit)
	r.HandleFunc("/srp/read/{pass_id}", srp.Get_ReadPassData)
	r.HandleFunc("/srp/update", srp.Post_UpdatePassData)
	r.HandleFunc("/srp/download/{pass_id}", srp.Get_DownloadPassData)
	r.HandleFunc("/srp/check/{pass_nr}/{regnum1}/{regnum2}/{regnum3}", srp.Get_CheckPass)

	// sector endpoints (7 z 7)
	r.HandleFunc("/sector/{sector_id}", sector.Initialize)
	r.HandleFunc("/sector/{sector_id}/states", sector.States)
	r.HandleFunc("/sector/{sector_id}/schedule", sector.Schedule)
	r.HandleFunc("/sector/notify/sendtosector/{rja_id}", sector.Notification_SendToSector)
	r.HandleFunc("/sector/notify/readytoleave/{rja_id}", sector.Notification_ReadyToLeave)
	r.HandleFunc("/sector/notify/onsector/{rja_id}", sector.Notification_OnSector)
	r.HandleFunc("/sector/notify/ontheroad/{rja_id}", sector.Notification_OnRoad)

	r.HandleFunc("/", RootHandler(host, port))

	// podłączenie routera do serwera HTTP
	http.Handle("/", r)
}
