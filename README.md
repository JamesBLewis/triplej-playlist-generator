# 🤖 triplej playlist generator ([live playlist](https://open.spotify.com/playlist/4wP3HpMngLebZ8pYvXD0Et?si=9f3747516a724f63))
Automatically generate a spotify playlist using the most recently played music on the [triplej radio station](https://www.abc.net.au/triplej).
## Getting Started
1. Register your application on [developer dashboard](https://developer.spotify.com/dashboard/applications) and obtain the `client_id` and a `client_secret`.
2. Use the autherization code on the offical spotify [web-api-auth-examples repo](https://github.com/spotify/web-api-auth-examples/tree/master/authorization_code) to obtain an `access_token` and a `refresh_token`. You will also need to modify the grants section of this logic to add a scope giving the app access to your public playlists. You can read about scopes [here](https://developer.spotify.com/documentation/general/guides/authorization/scopes/). As my playlist is public I only need the `playlist-modify-public` scope.
3. create a playlist in spotify and copy the link to it. Note we just want the playlist_id which is the bold section of the following example url:
h t t p s : / / open.spotify.com/playlist/**4wP3HpMngLebZ8pYvXD0Et**?si=9f3747516a724f63
4. edit main.go and replace the following with your credentials:
```
	spotifyClientId     = ""
	spotifyClientSecret = ""
	spotifyAccessToken  = ""
	spotifyRefreshToken = ""
	spotifyPlaylistId   = ""
```
5. simply run main.go
