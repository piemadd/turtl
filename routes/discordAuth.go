package routes

import (
	"encoding/json"
	"github.com/bwmarrin/discordgo"
	"github.com/dgrijalva/jwt-go"
	_ "github.com/joho/godotenv/autoload"
	"github.com/parnurzeal/gorequest"
	"net/http"
	"os"
	"strconv"
	"time"
	"turtl/db"
	"turtl/structs"
	"turtl/utils"
)

var discordAPIBase = "https://discord.com/api/v6"

func DiscordAuth(w http.ResponseWriter, r *http.Request) {
	// get url queries to extract code
	queries := r.URL.Query()
	code := queries.Get("code")

	// get actual user token
	request := gorequest.New()
	resp, body, errs := request.Post(discordAPIBase+"/oauth2/token").
		Set("Content-Type", "application/x-www-form-urlencoded").
		Send(structs.DiscordExchangeRequest{
			ClientId:     os.Getenv("DISCORD_CLIENT_ID"),
			ClientSecret: os.Getenv("DISCORD_CLIENT_SECRET"),
			GrantType:    "authorization_code",
			Code:         code,
			RedirectUri:  os.Getenv("APP_FQDN") + "/auth",
			Scope:        "identify",
		}).
		End()
	if errs != nil {
		_ = utils.HandleError(errs[0], "sending request to discord auth")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`Internal server error`))
		return
	}

	if resp.StatusCode != 200 {
		w.WriteHeader(resp.StatusCode)
		_, _ = w.Write([]byte(`Status code ` + strconv.Itoa(resp.StatusCode)))
		return
	}

	// put user info into a struct
	var discordAccount structs.DiscordExchangeResponse
	err := json.Unmarshal([]byte(body), &discordAccount)
	if utils.HandleError(err, "unmarshalling discord exchange") {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`Internal server error`))
		return
	}

	// get user info
	_, body, errs = request.Get(discordAPIBase+"/users/@me").
		Set("Authorization", discordAccount.TokenType+" "+discordAccount.AccessToken).
		End()
	if errs != nil {
		_ = utils.HandleError(errs[0], "sending request to get discord user")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`Internal server error`))
		revokeToken(discordAccount.AccessToken, w)
		return
	}

	var user discordgo.User
	err = json.Unmarshal([]byte(body), &user)
	if utils.HandleError(err, "unmarshalling discord user info") {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`Internal server error`))
		revokeToken(discordAccount.AccessToken, w)
		return
	}

	// get their turtl api key
	turtlAccount, ok := db.GetAccountFromDiscord(user.ID)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`Internal server error`))
		revokeToken(discordAccount.AccessToken, w)
		return
	}
	if turtlAccount.APIKey == "" {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`Account doesn't exist`))
		revokeToken(discordAccount.AccessToken, w)
		return
	}

	// create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"exp":    time.Now().Add(60 * time.Minute),
		"apikey": turtlAccount.APIKey,
	})

	tokenString, err := token.SignedString(utils.AppSecretKey)
	if utils.HandleError(err, "signing token") {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`Internal server error`))
		revokeToken(discordAccount.AccessToken, w)
		return
	}

	revokeToken(discordAccount.AccessToken, w)
	http.Redirect(w, r, os.Getenv("APP_FRONTEND_REDIRECT")+"?token="+tokenString, http.StatusFound)
}

func revokeToken(token string, w http.ResponseWriter) {
	_, _, errs := gorequest.New().Post(discordAPIBase+"/token/revoke").
		Set("Content-Type", "application/x-www-form-urlencoded").
		Send(`{"token":"` + token + `"}`).
		End()

	if errs != nil {
		_ = utils.HandleError(errs[0], "sending revocation request")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`Internal server error`))
		return
	}
}
