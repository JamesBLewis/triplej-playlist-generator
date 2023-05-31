.EXPORT_ALL_VARIABLES:
###########################
# add your config here
###########################
SPOTIFY_CLIENT_ID =
SPOTIFY_PLAYLIST_ID =
SPOTIFY_CLIENT_SECRET =
SPOTIFY_REFRESH_TOKEN =
PLAYLIST_SIZE = 30
##########################
bot:
	go run jobs/cmd/main.go

bot-container:
	docker build .
	docker run