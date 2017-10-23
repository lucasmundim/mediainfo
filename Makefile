run:
	GODEBUG=cgocheck=0 go run mediainfo.go --init fixtures/sample_video_1_init.mp4 --input=fixtures/sample_video_1_1.mp4