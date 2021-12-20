#!/bin/bash

set -e
if [ -z "$1" ]; then
  #判断用户输入
  read -p "Please choose all/proto/app: " -t 5 arg
else
  arg="$1"
fi

proto() {
  echo "proto"
  GOOS=$(go env GOOS)
  GOARCH=$(go env GOARCH)
  export PATH=$(pwd)/tools/$GOOS:$PATH

  # exPath="$(pwd)/tools/$GOOS"_"$GOARCH"
  # if [ -d "$exPath" ]; then
  #   export PATH="$exPath":$PATH
  # fi
  
  PROTO_SRC=./protocol/proto
  PROTO_DEST=./protocol/pb
  GENGO=go_out
  GENGRPC=go-grpc_out
  echo "gen proto ..."
  test -d ${PROTO_DEST} || mkdir -p ${PROTO_DEST}
  protoc -I=${PROTO_SRC} --${GENGO}=${PROTO_DEST} ${PROTO_SRC}/*.proto
  protoc -I=${PROTO_SRC} --${GENGRPC}=${PROTO_DEST} ${PROTO_SRC}/*.proto
  echo "gen proto ok"
}

build() {
  OUTPUT_DIR=$(pwd)/bin
  GOOS=$(go env GOOS)
  GO_VERSION=$(go version)
  major="1"
  minor="0"
  release=27

  APP_VERSION="${major}.${minor}.$release"
    if [ "$GOOS" = "darwin" ]; then
      sed -i "" "s/release=${release}/release=$((${release} + 1))/g" $0
    else
      sed -i "s/release=${release}/release=$((${release} + 1))/g" $0
    fi

  # test -f version.txt && APP_VERSION=$(cat version.txt)
  # array=(${APP_VERSION//./ })
  # array[2]=$((${array[2]} + 1))
  # APP_VERSION="${array[0]:-"1"}.${array[1]:-"0"}.${array[2]}"
  # echo ${APP_VERSION} >version.txt

  BUILD_VERSION=$(git log -1 --oneline)
  BUILD_TIME=$(date +%FT%T%z)
  GIT_REVISION=$(git rev-parse HEAD)
  GIT_BRANCH=$(git name-rev --name-only HEAD)

  echo "build version ${APP_VERSION}"

  for APP_NAME in "$@"; do
    echo "build ${APP_NAME}"
    go build -o ${OUTPUT_DIR}/\
      -gcflags "all=-N -l" \
      -ldflags "-s -X 'main.AppName=${APP_NAME}' \
            -X 'main.AppVersion=${APP_VERSION}' \
            -X 'main.BuildVersion=${BUILD_VERSION}' \
            -X 'main.BuildTime=${BUILD_TIME}' \
            -X 'main.GitRevision=${GIT_REVISION}' \
            -X 'main.GitBranch=${GIT_BRANCH}' \
            -X 'main.GoVersion=${GO_VERSION}'" \
      ./${APP_NAME}

  done
}

case $arg in
#判断变量cho的值
"all")
  proto
  build app
  ;;
"proto")
  proto
  ;;
"app")
  build app
  ;;
*)
  if [ -z "$arg" ]; then
    proto
    build app
  else
    echo "Your choose ($arg) is unknown!"
  fi

  ;;

esac
