task:
	go fmt 
	go mod tidy
	go build feed2cli.go input_standerd.go output_standerd.go sortableFeed.go input_uri.go output_slack.go merge.go diff.go
