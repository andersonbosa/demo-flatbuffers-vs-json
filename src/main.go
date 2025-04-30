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
	Id             uint64   `json:"id"`
	Name           string   `json:"name"`
	Age            uint     `json:"age"`
	Email          string   `json:"email"`
	Address        string   `json:"address"`
	Phone          string   `json:"phone"`
	CreatedAt      string   `json:"created_at"`
	UpdatedAt      string   `json:"updated_at"`
	IsActive       bool     `json:"is_active"`
	LastLogin      string   `json:"last_login"`
	ProfilePicture string   `json:"profile_picture"`
	Roles          []string `json:"roles"`
}

func generateUserData(n int) ([]UserJson, [][]byte) {
	data := make([]UserJson, 0, n)
	names := make([][]byte, 0, n)
	now := time.Now()

	for i := 0; i < n; i++ {
		user := UserJson{
			Id:             uint64(i),
			Name:           fmt.Sprintf("User_%d", i),
			Age:            uint(i % 100),
			Email:          fmt.Sprintf("user%d@example.com", i),
			Address:        fmt.Sprintf("Rua %d", i),
			Phone:          fmt.Sprintf("+551199%06d", i),
			CreatedAt:      now.Add(-24 * time.Hour).Format(time.RFC3339),
			UpdatedAt:      now.Format(time.RFC3339),
			IsActive:       i%2 == 0,
			LastLogin:      now.Add(-time.Duration(i) * time.Minute).Format(time.RFC3339),
			ProfilePicture: fmt.Sprintf("https://avatar.iran.liara.run/username?username=%d", i),
			Roles:          []string{"user", "viewer"},
		}
		data = append(data, user)
		names = append(names, []byte(user.Name))
	}
	return data, names
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

func benchmarkFlatBuffers(usersData []UserJson) ([]byte, time.Duration, time.Duration) {
	builder := flatbuffers.NewBuilder(0)
	offsets := make([]flatbuffers.UOffsetT, len(usersData))

	startBuild := time.Now()

	for i := len(usersData) - 1; i >= 0; i-- {
		u := usersData[i]

		name := builder.CreateString(u.Name)
		email := builder.CreateString(u.Email)
		address := builder.CreateString(u.Address)
		phone := builder.CreateString(u.Phone)
		createdAt := builder.CreateString(u.CreatedAt)
		updatedAt := builder.CreateString(u.UpdatedAt)
		lastLogin := builder.CreateString(u.LastLogin)
		profilePic := builder.CreateString(u.ProfilePicture)

		roleOffsets := make([]flatbuffers.UOffsetT, len(u.Roles))
		for j := len(u.Roles) - 1; j >= 0; j-- {
			roleOffsets[j] = builder.CreateString(u.Roles[j])
		}
		users.UserStartRolesVector(builder, len(roleOffsets))
		for j := len(roleOffsets) - 1; j >= 0; j-- {
			builder.PrependUOffsetT(roleOffsets[j])
		}
		roles := builder.EndVector(len(roleOffsets))

		users.UserStart(builder)
		users.UserAddId(builder, u.Id)
		users.UserAddName(builder, name)
		users.UserAddAge(builder, uint32(u.Age))
		users.UserAddEmail(builder, email)
		users.UserAddAddress(builder, address)
		users.UserAddPhone(builder, phone)
		users.UserAddCreatedAt(builder, createdAt)
		users.UserAddUpdatedAt(builder, updatedAt)
		users.UserAddIsActive(builder, u.IsActive)
		users.UserAddLastLogin(builder, lastLogin)
		users.UserAddProfilePicture(builder, profilePic)
		users.UserAddRoles(builder, roles)
		offsets[i] = users.UserEnd(builder)
	}

	users.UserListStartUsersVector(builder, len(offsets))
	for i := len(offsets) - 1; i >= 0; i-- {
		builder.PrependUOffsetT(offsets[i])
	}
	userVec := builder.EndVector(len(offsets))

	users.UserListStart(builder)
	users.UserListAddUsers(builder, userVec)
	root := users.UserListEnd(builder)
	builder.Finish(root)
	buildTime := time.Since(startBuild)

	buf := builder.FinishedBytes()

	startRead := time.Now()
	list := users.GetRootAsUserList(buf, 0)
	var user users.User
	for i := 0; i < list.UsersLength(); i++ {
		list.Users(&user, i)
		_ = user.Id()
		_ = user.Name()
		_ = user.Email()
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
	jsonData, _ := generateUserData(dataSize)

	log.Println("Running JSON benchmark...")
	jsonBuf, jsonWriteTime, jsonReadTime := benchmarkJson(jsonData)
	log.Printf("JSON -> Size: %.3f MB | Write: %v | Read: %v\n", bytesToMB(len(jsonBuf)), jsonWriteTime, jsonReadTime)

	log.Println("Running FlatBuffers benchmark...")
	flatBuf, flatWriteTime, flatReadTime := benchmarkFlatBuffers(jsonData)
	log.Printf("FlatBuffers -> Size: %.3f MB | Write: %v | Read: %v\n", bytesToMB(len(flatBuf)), flatWriteTime, flatReadTime)
}
