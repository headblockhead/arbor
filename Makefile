build:
	docker build -t headblockhead/arbor/display:0.1 .
display: build
	docker run -p 6868:6868 -it headblockhead/arbor/display:0.1