docker build -t random-dev -f docker/Dockerfile-dev
ocker run --rm -it -p 8000:3000 -v $(PWD):/go/src/app --network dev random-dev

