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
	"gokongres/api/ws"
	"gokongres/sessions"
)

// RegisterHandlers rejestruje endpointy HTTP używane przez serwera.
func RegisterHandlers(host string, port int) {
	r := http.NewServeMux()

	// Middleware sesji (nagłówki no-cache) jest podłączany przy montowaniu routera
	// na końcu tej funkcji - http.ServeMux nie wspiera metody .Use().

	// protect opakowuje handler wymaganiem aktywnej sesji (odpowiednik @login_required
	// ze starego API). Sesję zakładają: panel admina (/auth/admin), zbory (/auth/login)
	// oraz działy PK (/pk/login). Endpointy publiczne (config, ekrany monitoring/sektor/
	// bufor/przyjazdy/czw, skanery */check, listy potrzebne przed logowaniem jak /pk/hints
	// i odczyty /rja używane przez publiczny wyświetlacz rozkładu) pozostają bez ochrony.
	protect := func(h http.HandlerFunc) http.Handler {
		return sessions.AuthMiddleware(h)
	}

	// config endpoints
	r.HandleFunc("GET /config/all", config.Get_AllConfig) // zamiast /config/tury
	r.HandleFunc("GET /config/active/tura", config.Get_ActiveTura)

	// auth endpoints
	r.HandleFunc("POST /auth/login", auth.LoginHandler)
	r.HandleFunc("POST /auth/admin", auth.AdminHandler)
	r.Handle("POST /auth/logout", protect(auth.LogoutHandler))
	r.HandleFunc("POST /auth/permissions", auth.PermissionsHandler)

	// arrivals endpoints
	r.HandleFunc("GET /arrivals/all", arrivals.Get_All)
	r.HandleFunc("POST /arrivals/set", arrivals.Set)

	// limits endpoints (panel admina - wymaga sesji; w starym API brak ochrony, świadomie zaostrzone)
	r.Handle("GET /limits/zbory", protect(limits.Get_Zbory))
	r.Handle("POST /limits/zbory/setlimit", protect(limits.Post_SetZboryLimit))
	r.Handle("GET /limits/dzialy", protect(limits.Get_Dzialy))
	r.Handle("POST /limits/dzialy/setlimit", protect(limits.Post_SetDzialyLimit))

	// PK (parking księżycowy - działy) endpoints
	// publiczne: login (wejście, zakłada sesję), hints (lista działów przed logowaniem), check (skaner)
	r.HandleFunc("POST /pk/login", pk.Get_Login)
	r.HandleFunc("GET /pk/hints", pk.Get_Hints)
	r.HandleFunc("GET /pk/check", pk.Get_CheckPass)
	r.HandleFunc("GET /pk/free", pk.Get_FreePass)
	r.Handle("GET /pk/all", protect(pk.Get_LoadAll))
	r.Handle("POST /pk/create", protect(pk.Get_CreatePassID))
	r.Handle("POST /pk/find", protect(pk.Get_FindPassID))
	r.Handle("GET /pk/read/{pk_id}", protect(pk.Get_ReadPassData))
	r.Handle("POST /pk/update", protect(pk.Post_UpdatePassData))
	r.Handle("GET /pk/delete/{pk_id}", protect(pk.Get_DeletePass))
	r.Handle("GET /pk/download/{pk_id}", protect(pk.Get_DownloadPassData))
	r.Handle("GET /pk/isfreepass/{dep_name}/{tura}", protect(pk.Get_IsFreePass))

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

	// ia endpoints (portal zboru - wymaga zalogowania)
	r.Handle("GET /ia/list/{congregation_name}", protect(ia.Get_List))
	r.Handle("GET /ia/download/{sra_id}", protect(ia.Get_Download))

	// rja endpoints
	r.HandleFunc("GET /rja/zbory/{tura_id}", rja.Get_CongregationList)
	r.HandleFunc("GET /rja/zbor/{congregation_id}", rja.Get_CongregationRJA)
	r.HandleFunc("GET /rja/sra/{tura_id}", rja.Get_SraList)
	r.HandleFunc("GET /rja/terminals", rja.Get_TerminalsList)
	r.HandleFunc("GET /rja/sectors/{terminal_id}", rja.Get_SectorsList)
	r.HandleFunc("GET /rja/buses/{sector_id}/{tura_id}", rja.Get_BusesOfSector)
	// odczyty RJA pozostają publiczne (używa ich publiczny wyświetlacz rozkładu bez logowania);
	// chronione są tylko operacje panelu admina
	r.Handle("GET /rja/buses/used/{tura_id}", protect(rja.Get_BusesUsed))
	r.Handle("POST /rja/buses/save", protect(rja.Get_BusesSave))

	// sra endpoints (search/congregations publiczne - używane przy wyborze zboru przed logowaniem)
	r.HandleFunc("GET /sra/search/congregations/{pattern}", sra.Get_SearchCongregationsByPattern)
	r.Handle("POST /sra/submit/bus", protect(sra.Post_SubmitBus))
	r.Handle("PUT /sra/submit/nobus/{congregation_name}", protect(sra.Put_SubmitNoBus))
	r.Handle("POST /sra/check_pilot_duplicate", protect(sra.Post_IsPilotDuplicate))
	r.Handle("GET /sra/table", protect(sra.Get_Table))
	r.Handle("POST /sra/save", protect(sra.Post_Save))
	r.Handle("GET /sra/delete/{sra_id}", protect(sra.Get_Delete))
	r.Handle("GET /sra/export/xlsx", protect(sra.Get_Table_Export_Xlsx))

	// czw (wydawanie zastępczych identyfikatorów parkingowych) endpoints
	r.HandleFunc("GET /czw/init", czw.Get_Init)
	r.HandleFunc("POST /czw/issuing", czw.Post_Issuing)
	r.HandleFunc("POST /czw/search", czw.Post_Search)
	r.HandleFunc("POST /czw/cancellation", czw.Post_Cancellation)

	// srp endpoints (check publiczny - skaner; pozostałe wymagają sesji zboru/admina)
	r.HandleFunc("GET /srp/check", srp.Get_CheckPass)
	r.HandleFunc("GET /srp/free", srp.Get_FreePass)
	r.Handle("GET /srp/zbory", protect(srp.Get_CongregationList))
	r.Handle("GET /srp/all", protect(srp.Get_AllList))
	r.Handle("POST /srp/create", protect(srp.Post_Create))
	r.Handle("POST /srp/find", protect(srp.Post_FindPassID))
	r.Handle("GET /srp/delete/{srp_id}", protect(srp.Get_Delete))
	r.Handle("GET /srp/isfreepass/{congregation_name}", protect(srp.Get_IsFreePass))
	r.Handle("GET /srp/limit/{congregation_name}", protect(srp.Get_UsingLimit))
	r.Handle("POST /srp/limit/change", protect(srp.Post_RequestNewLimit))
	r.Handle("GET /srp/read/{pass_id}", protect(srp.Get_ReadPassData))
	r.Handle("POST /srp/update", protect(srp.Post_UpdatePassData))
	r.Handle("GET /srp/download/{pass_id}", protect(srp.Get_DownloadPassData))

	// sector endpoints
	r.HandleFunc("GET /sector/{sector_id}", sector.Initialize)
	r.HandleFunc("GET /sector/{sector_id}/states", sector.States)
	r.HandleFunc("GET /sector/{sector_id}/schedule", sector.Schedule)
	r.HandleFunc("GET /sector/notify/sendtosector/{rja_id}", sector.Notification_SendToSector)
	r.HandleFunc("GET /sector/notify/readytoleave/{rja_id}", sector.Notification_ReadyToLeave)
	r.HandleFunc("GET /sector/notify/onsector/{rja_id}", sector.Notification_OnSector)
	r.HandleFunc("GET /sector/notify/ontheroad/{rja_id}", sector.Notification_OnRoad)

	// websocket — powiadamianie o zmianach stanów odprawy autokarów (publiczny, jak /states i /notify/*)
	r.HandleFunc("GET /ws/odprawa", ws.HandleWS)

	r.HandleFunc("/", RootHandler(host, port))

	// podłączenie routera do serwera HTTP wraz z middleware sesji (nagłówki no-cache),
	// dzięki czemu odpowiedzi panelu admina (np. /auth/permissions) nie są cache'owane
	http.Handle("/", sessions.SessionMiddleware(r))
}
