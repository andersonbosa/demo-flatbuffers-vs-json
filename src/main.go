package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/andersonbosa/demo-flatbuffers-vs-json/users"
	flatbuffers "github.com/google/flatbuffers/go"
)

type UserJson struct {
	Name string `json:"name"`
	Id   uint64 `json:"id"`
}

func generateUserData(n int) ([]UserJson, [][]byte) {
	jsonData := make([]UserJson, 0, n)
	names := make([][]byte, 0, n)
	for i := 0; i < n; i++ {
		name := fmt.Sprintf("User_%d", i)
		jsonData = append(jsonData, UserJson{Name: name, Id: uint64(i)})
		names = append(names, []byte(name))
	}
	return jsonData, names
}

func benchmarkJson(users []UserJson) ([]byte, time.Duration, time.Duration) {
	startMarshal := time.Now()
	buf, err := json.Marshal(users)
	if err != nil {
		log.Fatal(err)
	}
	marshalTime := time.Since(startMarshal)

	startUnmarshal := time.Now()
	var decoded []UserJson
	if err := json.Unmarshal(buf, &decoded); err != nil {
		log.Fatal(err)
	}
	unmarshalTime := time.Since(startUnmarshal)

	return buf, marshalTime, unmarshalTime
}

func benchmarkFlatBuffers(names [][]byte) ([]byte, time.Duration, time.Duration) {
	builder := flatbuffers.NewBuilder(0)
	offsets := make([]flatbuffers.UOffsetT, len(names))

	startBuild := time.Now()
	for i, name := range names {
		nameOffset := builder.CreateByteString(name)
		users.UserStart(builder)
		users.UserAddName(builder, nameOffset)
		users.UserAddId(builder, uint64(i))
		offsets[i] = users.UserEnd(builder)
	}
	users.UserListStartUsersVector(builder, len(offsets))
	for i := len(offsets) - 1; i >= 0; i-- {
		builder.PrependUOffsetT(offsets[i])
	}
	usersVector := builder.EndVector(len(offsets))

	users.UserListStart(builder)
	users.UserListAddUsers(builder, usersVector)
	root := users.UserListEnd(builder)
	builder.Finish(root)
	buildTime := time.Since(startBuild)

	buf := builder.FinishedBytes()

	startRead := time.Now()
	list := users.GetRootAsUserList(buf, 0)
	for i := 0; i < list.UsersLength(); i++ {
		var user users.User
		list.Users(&user, i)
		_ = user.Name()
		_ = user.Id()
	}
	readTime := time.Since(startRead)

	return buf, buildTime, readTime
}

func bytesToMB(bytes int) float64 {
	return float64(bytes) / 1024 / 1024
}
func main() {
	log.Println("Benchmarking JSON vs FlatBuffers...")

	const dataSize = 1_000_000
	jsonData, names := generateUserData(dataSize)

	log.Println("Running JSON benchmark...")
	jsonBuf, jsonWriteTime, jsonReadTime := benchmarkJson(jsonData)
	log.Printf("JSON -> Size: %.3f MB | Write: %v | Read: %v\n", bytesToMB(len(jsonBuf)), jsonWriteTime, jsonReadTime)

	log.Println("Running FlatBuffers benchmark...")
	flatBuf, flatWriteTime, flatReadTime := benchmarkFlatBuffers(names)
	log.Printf("FlatBuffers -> Size: %.3f MB | Write: %v | Read: %v\n", bytesToMB(len(flatBuf)), flatWriteTime, flatReadTime)
}
