package api

import (
	"net/http"

	"gokongres/api/arrivals"
	"gokongres/api/auth"
	"gokongres/api/buffer"
	"gokongres/api/config"
	"gokongres/api/czw"
	"gokongres/api/ia"
	"gokongres/api/limits"
	"gokongres/api/monitoring"
	"gokongres/api/pk"
	"gokongres/api/rja"
	"gokongres/api/sector"
	"gokongres/api/sra"
	"gokongres/api/srp"
	"gokongres/api/terminals"
	"gokongres/api/users"
	"gokongres/sessions"
)

// RegisterHandlers rejestruje endpointy HTTP używane przez serwera.
func RegisterHandlers(host string, port int) {
	r := http.NewServeMux()

	// Middleware sesji (nagłówki no-cache) jest podłączany przy montowaniu routera
	// na końcu tej funkcji - http.ServeMux nie wspiera metody .Use().

	// config endpoints
	r.HandleFunc("GET /config/all", config.Get_AllConfig) // zamiast /config/tury
	r.HandleFunc("GET /config/active/tura", config.Get_ActiveTura)

	// auth endpoints
	r.HandleFunc("POST /auth/login", auth.LoginHandler)
	r.HandleFunc("POST /auth/admin", auth.AdminHandler)
	r.HandleFunc("POST /auth/logout", auth.LogoutHandler)
	r.HandleFunc("POST /auth/permissions", auth.PermissionsHandler)

	// arrivals endpoints
	r.HandleFunc("GET /arrivals/all", arrivals.Get_All)
	r.HandleFunc("POST /arrivals/set", arrivals.Set)

	// limits endpoints
	r.HandleFunc("GET /limits/zbory", limits.Get_Zbory)
	r.HandleFunc("POST /limits/zbory/setlimit", limits.Post_SetZboryLimit)
	r.HandleFunc("GET /limits/dzialy", limits.Get_Dzialy)
	r.HandleFunc("POST /limits/dzialy/setlimit", limits.Post_SetDzialyLimit)

	// PK (parking księżycowy - działy) endpoints
	r.HandleFunc("GET /pk/hints", pk.Get_Hints)
	r.HandleFunc("POST /pk/login", pk.Get_Login)
	r.HandleFunc("GET /pk/all", pk.Get_LoadAll)
	r.HandleFunc("POST /pk/create", pk.Get_CreatePassID)
	r.HandleFunc("POST /pk/find", pk.Get_FindPassID)
	r.HandleFunc("GET /pk/read/{pk_id}", pk.Get_ReadPassData)
	r.HandleFunc("POST /pk/update", pk.Post_UpdatePassData)
	r.HandleFunc("GET /pk/delete/{pk_id}", pk.Get_DeletePass)
	r.HandleFunc("GET /pk/download/{pk_id}", pk.Get_DownloadPassData)
	r.HandleFunc("GET /pk/isfreepass/{dep_name}/{tura}", pk.Get_IsFreePass)
	r.HandleFunc("GET /pk/check", pk.Get_CheckPass)

	// monitoring endpoints
	r.HandleFunc("GET /monitoring/states", monitoring.Get_StatesRepo)
	r.HandleFunc("GET /monitoring/terminals", monitoring.Get_TerminalsList)

	// buffer endpoints
	r.HandleFunc("GET /buffer/all", buffer.Get_AllShortList)
	r.HandleFunc("GET /buffer/fullinfo/{terminal_name}", buffer.Get_FullInfo)
	r.HandleFunc("GET /buffer/states/{terminal_name}", buffer.Get_States)
	r.HandleFunc("GET /buffer/notify/nobus/{rja_id}", buffer.Get_NoBusNotification)
	r.HandleFunc("GET /buffer/notify/inbuffer/{rja_id}", buffer.Get_InBufferNotification)
	r.HandleFunc("GET /buffer/notify/secondcircle/{rja_id}", buffer.Get_SecondCircleNotification)
	r.HandleFunc("GET /buffer/notify/sendtosector/{rja_id}", buffer.Get_SendToSectorNotification)

	// terminals endpoints
	r.HandleFunc("GET /terminals/all", terminals.Get_AllList)
	r.HandleFunc("GET /terminals/fullinfo/{terminal_id}", terminals.Get_FullInfo)

	// users (konta admin) endpoints — chronione (is_users)
	r.HandleFunc("GET /users/all", users.Get_All)
	r.HandleFunc("POST /users/create", users.Post_Create)
	r.HandleFunc("POST /users/update", users.Post_Update)
	r.HandleFunc("GET /users/delete/{id}", users.Get_Delete)

	// ia endpoints
	r.HandleFunc("GET /ia/list/{congregation_name}", ia.Get_List)
	r.HandleFunc("GET /ia/download/{sra_id}", ia.Get_Download)

	// rja endpoints
	r.HandleFunc("GET /rja/zbory/{tura_id}", rja.Get_CongregationList)
	r.HandleFunc("GET /rja/zbor/{congregation_id}", rja.Get_CongregationRJA)
	r.HandleFunc("GET /rja/sra/{tura_id}", rja.Get_SraList)
	r.HandleFunc("GET /rja/terminals", rja.Get_TerminalsList)
	r.HandleFunc("GET /rja/sectors/{terminal_id}", rja.Get_SectorsList)
	r.HandleFunc("GET /rja/buses/{sector_id}/{tura_id}", rja.Get_BusesOfSector)
	r.HandleFunc("GET /rja/buses/used/{tura_id}", rja.Get_BusesUsed)
	r.HandleFunc("POST /rja/buses/save", rja.Get_BusesSave)

	// sra endpoints
	r.HandleFunc("GET /sra/search/congregations/{pattern}", sra.Get_SearchCongregationsByPattern)
	r.HandleFunc("POST /sra/submit/bus", sra.Post_SubmitBus)
	r.HandleFunc("PUT /sra/submit/nobus/{congregation_name}", sra.Put_SubmitNoBus)
	r.HandleFunc("POST /sra/check_pilot_duplicate", sra.Post_IsPilotDuplicate)
	r.HandleFunc("GET /sra/table", sra.Get_Table)
	r.HandleFunc("POST /sra/save", sra.Post_Save)
	r.HandleFunc("GET /sra/delete/{sra_id}", sra.Get_Delete)
	r.HandleFunc("GET /sra/export/xlsx", sra.Get_Table_Export_Xlsx)

	// czw (wydawanie zastępczych identyfikatorów parkingowych) endpoints
	r.HandleFunc("GET /czw/init", czw.Get_Init)
	r.HandleFunc("POST /czw/issuing", czw.Post_Issuing)
	r.HandleFunc("POST /czw/search", czw.Post_Search)
	r.HandleFunc("POST /czw/cancellation", czw.Post_Cancellation)

	// srp endpoints
	r.HandleFunc("GET /srp/zbory", srp.Get_CongregationList)
	r.HandleFunc("GET /srp/all", srp.Get_AllList)
	r.HandleFunc("POST /srp/create", srp.Post_Create)
	r.HandleFunc("POST /srp/find", srp.Post_FindPassID)
	r.HandleFunc("GET /srp/delete/{srp_id}", srp.Get_Delete)
	r.HandleFunc("GET /srp/isfreepass/{congregation_name}", srp.Get_IsFreePass)
	r.HandleFunc("GET /srp/limit/{congregation_name}", srp.Get_UsingLimit)
	r.HandleFunc("POST /srp/limit/change", srp.Post_RequestNewLimit)
	r.HandleFunc("GET /srp/read/{pass_id}", srp.Get_ReadPassData)
	r.HandleFunc("POST /srp/update", srp.Post_UpdatePassData)
	r.HandleFunc("GET /srp/download/{pass_id}", srp.Get_DownloadPassData)
	r.HandleFunc("GET /srp/check", srp.Get_CheckPass)

	// sector endpoints
	r.HandleFunc("GET /sector/{sector_id}", sector.Initialize)
	r.HandleFunc("GET /sector/{sector_id}/states", sector.States)
	r.HandleFunc("GET /sector/{sector_id}/schedule", sector.Schedule)
	r.HandleFunc("GET /sector/notify/sendtosector/{rja_id}", sector.Notification_SendToSector)
	r.HandleFunc("GET /sector/notify/readytoleave/{rja_id}", sector.Notification_ReadyToLeave)
	r.HandleFunc("GET /sector/notify/onsector/{rja_id}", sector.Notification_OnSector)
	r.HandleFunc("GET /sector/notify/ontheroad/{rja_id}", sector.Notification_OnRoad)

	r.HandleFunc("/", RootHandler(host, port))

	// podłączenie routera do serwera HTTP wraz z middleware sesji (nagłówki no-cache),
	// dzięki czemu odpowiedzi panelu admina (np. /auth/permissions) nie są cache'owane
	http.Handle("/", sessions.SessionMiddleware(r))
}
