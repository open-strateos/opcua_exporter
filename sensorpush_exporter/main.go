package main

import (
	"context"
	"log"
	"sensorpush_exporter/sensorpush"
)

func main() {
	username := "internal-tools@transcriptic.com"
	password := "7ranscripticPush!"
	client := getClient()
	authCtx := getAuthContext(client, username, password)

	sensorNameMap := getSensorNameMap(authCtx, client)
	getSamples(authCtx, client, *sensorNameMap)
}

func getClient() *sensorpush.APIClient {
	config := sensorpush.NewConfiguration()
	client := sensorpush.NewAPIClient(config)

	return client
}

func getAuthContext(client *sensorpush.APIClient, username string, password string) context.Context {
	ctx := context.Background()
	authResp, _, err := client.ApiApi.OauthAuthorizePost(ctx, sensorpush.AuthorizeRequest{
		Email:    username,
		Password: password,
	})
	token, _, err := client.ApiApi.AccessToken(ctx, sensorpush.AccessTokenRequest{
		Authorization: authResp.Authorization,
	})
	if err != nil {
		log.Fatal(err)
	}

	authCtx := context.WithValue(ctx, sensorpush.ContextAccessToken, token.Accesstoken)

	return authCtx
}

type SensorNameMap map[string]string

func getSensorNameMap(authCtx context.Context, client *sensorpush.APIClient) *SensorNameMap {
	sensors, _, err := client.ApiApi.Sensors(authCtx, sensorpush.SensorsRequest{})
	if err != nil {
		log.Fatal(err)
	}

	nameMap := make(SensorNameMap)
	for _, v := range sensors {
		nameMap[v.Id] = v.Name
	}
	return &nameMap
}

func getSamples(authCtx context.Context, client *sensorpush.APIClient, sensorNameMap SensorNameMap) map[string]sensorpush.Sample {

	samples, resp, err := client.ApiApi.Samples(authCtx, sensorpush.SamplesRequest{
		Limit: 1,
	})
	if err != nil {
		log.Print("CODE: ", resp.StatusCode)
		log.Fatal(err)
	}

	result := make(map[string]sensorpush.Sample)
	for sensorId, samples := range samples.Sensors {
		sensorName := sensorNameMap[sensorId]
		result[sensorName] = samples[0]
	}
	return result

}
