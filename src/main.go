package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/andersonbosa/demo-flatbuffers-vs-json/users"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
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

func generateUserData(n int) []UserJson {
	data := make([]UserJson, 0, n)
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
			ProfilePicture: fmt.Sprintf("https://img.com/%d.jpg", i),
			Roles:          []string{"user", "viewer"},
		}
		data = append(data, user)
	}
	return data
}

func writeFile(filename string, data []byte) {
	err := os.WriteFile(filename, data, 0644)
	if err != nil {
		log.Fatalf("Erro ao escrever %s: %v", filename, err)
	}
	log.Printf("Arquivo %s salvo com sucesso (%.2f MB)", filename, bytesToMB(len(data)))
}

func benchmarkJsonWrite(users []UserJson) ([]byte, time.Duration) {
	start := time.Now()
	buf, err := json.Marshal(users)
	if err != nil {
		log.Fatal(err)
	}
	return buf, time.Since(start)
}

func benchmarkJsonRead(buf []byte) time.Duration {
	start := time.Now()
	var users []UserJson
	err := json.Unmarshal(buf, &users)
	if err != nil {
		log.Fatal(err)
	}
	return time.Since(start)
}

func benchmarkFlatBuffersWrite(usersData []UserJson) ([]byte, time.Duration) {
	builder := flatbuffers.NewBuilder(0)
	offsets := make([]flatbuffers.UOffsetT, len(usersData))

	start := time.Now()
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
	usersVec := builder.EndVector(len(offsets))

	users.UserListStart(builder)
	users.UserListAddUsers(builder, usersVec)
	root := users.UserListEnd(builder)
	builder.Finish(root)

	return builder.FinishedBytes(), time.Since(start)
}

func benchmarkFlatBuffersRead(buf []byte) time.Duration {
	start := time.Now()
	list := users.GetRootAsUserList(buf, 0)
	var user users.User
	for i := 0; i < list.UsersLength(); i++ {
		list.Users(&user, i)
		_ = user.Id()
		_ = user.Name()
	}
	return time.Since(start)
}

func bytesToMB(bytes int) float64 {
	return float64(bytes) / 1024.0 / 1024.0
}

func generateBarItems(values []float64) []opts.BarData {
	items := make([]opts.BarData, len(values))
	for i, v := range values {
		items[i] = opts.BarData{Value: v}
	}
	return items
}

func renderCharts(jsonTimes, flatTimes []float64, jsonSizeMB, flatSizeMB float64) {
	barTime := charts.NewBar()
	barTime.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "Benchmark Tempo", Subtitle: "JSON vs FlatBuffers"}),
		charts.WithYAxisOpts(opts.YAxis{Name: "Tempo (ms)"}),
	)
	barTime.SetXAxis([]string{"Write", "Read"}).
		AddSeries("JSON", generateBarItems(jsonTimes)).
		AddSeries("FlatBuffers", generateBarItems(flatTimes))
	f1, _ := os.Create("benchmark_tempo_chart.html")
	defer f1.Close()
	barTime.Render(f1)

	barSize := charts.NewBar()
	barSize.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "Benchmark Tamanho", Subtitle: "JSON vs FlatBuffers"}),
		charts.WithYAxisOpts(opts.YAxis{Name: "Tamanho (MB)"}),
	)
	barSize.SetXAxis([]string{"JSON", "FlatBuffers"}).
		AddSeries("Tamanho", generateBarItems([]float64{jsonSizeMB, flatSizeMB}))
	f2, _ := os.Create("benchmark_tamanho_chart.html")
	defer f2.Close()
	barSize.Render(f2)
}

func main() {
	const dataSize = 1_000_000
	log.Println("Gerando dados...")
	users := generateUserData(dataSize)

	log.Println("Serializando JSON...")
	jsonBuf, jsonWrite := benchmarkJsonWrite(users)
	writeFile("output.json", jsonBuf)

	log.Println("Serializando FlatBuffers...")
	flatBuf, flatWrite := benchmarkFlatBuffersWrite(users)
	writeFile("output.bin", flatBuf)

	log.Println("Lendo JSON...")
	jsonRead := benchmarkJsonRead(jsonBuf)

	log.Println("Lendo FlatBuffers...")
	flatRead := benchmarkFlatBuffersRead(flatBuf)

	jsonSizeMB := bytesToMB(len(jsonBuf))
	flatSizeMB := bytesToMB(len(flatBuf))

	log.Println("Resultados:")
	log.Printf("JSON        -> %.2f MB | Write: %v | Read: %v", jsonSizeMB, jsonWrite, jsonRead)
	log.Printf("FlatBuffers -> %.2f MB | Write: %v | Read: %v", flatSizeMB, flatWrite, flatRead)

	renderCharts(
		[]float64{float64(jsonWrite.Milliseconds()), float64(jsonRead.Milliseconds())},
		[]float64{float64(flatWrite.Milliseconds()), float64(flatRead.Milliseconds())},
		jsonSizeMB,
		flatSizeMB,
	)
	log.Println("Charts gerados.")
}
