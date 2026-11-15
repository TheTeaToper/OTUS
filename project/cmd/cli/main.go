package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

const (
	defaultBaseURL = "http://localhost:8080/api/v1"
	keySubnet      = "subnet"
)

var (
	baseURL string
	login   string
	ip      string
	subnet  string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "ab-admin",
		Short: "CLI утилита для администрирования сервиса Анти-брутфорс",
	}

	// Глобальный флаг для изменения адреса сервера при необходимости
	rootCmd.PersistentFlags().StringVar(&baseURL, "url", defaultBaseURL, "Базовый URL API сервиса")

	// Команда сброса бакетов
	resetCmd := &cobra.Command{
		Use:   "reset",
		Short: "Сбросить бакеты для указанного логина и IP",
		Run:   runReset,
	}
	resetCmd.Flags().StringVar(&login, "login", "", "Логин для сброса (обязательно)")
	resetCmd.Flags().StringVar(&ip, "ip", "", "IP для сброса (обязательно)")
	_ = resetCmd.MarkFlagRequired("login")
	_ = resetCmd.MarkFlagRequired("ip")

	// Команды для Whitelist
	whitelistCmd := &cobra.Command{Use: "whitelist", Short: "Управление белым списком"}
	addWhiteCmd := &cobra.Command{Use: "add", Short: "Добавить подсеть в whitelist", Run: runAddWhitelist}
	removeWhiteCmd := &cobra.Command{Use: "remove", Short: "Удалить подсеть из whitelist", Run: runRemoveWhitelist}
	addWhiteCmd.Flags().StringVar(&subnet, "subnet", "", "Подсеть в формате IP/Маска (например, 192.1.1.0/25)")
	removeWhiteCmd.Flags().StringVar(&subnet, "subnet", "", "Подсеть в формате IP/Маска (например, 192.1.1.0/25)")
	_ = addWhiteCmd.MarkFlagRequired("subnet")
	_ = removeWhiteCmd.MarkFlagRequired("subnet")
	whitelistCmd.AddCommand(addWhiteCmd, removeWhiteCmd)

	// Команды для Blacklist
	blacklistCmd := &cobra.Command{Use: "blacklist", Short: "Управление черным списком"}
	addBlackCmd := &cobra.Command{Use: "add", Short: "Добавить подсеть в blacklist", Run: runAddBlacklist}
	removeBlackCmd := &cobra.Command{Use: "remove", Short: "Удалить подсеть из blacklist", Run: runRemoveBlacklist}
	addBlackCmd.Flags().StringVar(&subnet, "subnet", "", "Подсеть в формате IP/Маска (например, 192.1.1.0/25)")
	removeBlackCmd.Flags().StringVar(&subnet, "subnet", "", "Подсеть в формате IP/Маска (например, 192.1.1.0/25)")
	_ = addBlackCmd.MarkFlagRequired("subnet")
	_ = removeBlackCmd.MarkFlagRequired("subnet")
	blacklistCmd.AddCommand(addBlackCmd, removeBlackCmd)

	// Объединяем команды
	rootCmd.AddCommand(resetCmd, whitelistCmd, blacklistCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// --- Обработчики команд ---

func runReset(cmd *cobra.Command, args []string) {
	body, _ := json.Marshal(map[string]string{"login": login, "ip": ip})
	sendRequest(http.MethodPost, baseURL+"/buckets/reset", body, "Бакеты успешно сброшены")
}

func runAddWhitelist(cmd *cobra.Command, args []string) {
	body, _ := json.Marshal(map[string]string{keySubnet: subnet})
	sendRequest(http.MethodPost, baseURL+"/whitelist", body, fmt.Sprintf("Подсеть %s добавлена в whitelist", subnet))
}

func runRemoveWhitelist(cmd *cobra.Command, args []string) {
	body, _ := json.Marshal(map[string]string{keySubnet: subnet})
	sendRequest(http.MethodDelete, baseURL+"/whitelist", body, fmt.Sprintf("Подсеть %s удалена из whitelist", subnet))
}

func runAddBlacklist(cmd *cobra.Command, args []string) {
	body, _ := json.Marshal(map[string]string{keySubnet: subnet})
	sendRequest(http.MethodPost, baseURL+"/blacklist", body, fmt.Sprintf("Подсеть %s добавлена в blacklist", subnet))
}

func runRemoveBlacklist(cmd *cobra.Command, args []string) {
	body, _ := json.Marshal(map[string]string{keySubnet: subnet})
	sendRequest(http.MethodDelete, baseURL+"/blacklist", body, fmt.Sprintf("Подсеть %s удалена из blacklist", subnet))
}

// Вспомогательная функция отправки запросов на сервер.
func sendRequest(method, url string, body []byte, successMsg string) {
	req, err := http.NewRequestWithContext(context.Background(), method, url, bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("❌ Ошибка создания запроса: %v\n", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("❌ Сервер недоступен: %v. Сначала запустите основной сервис.\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Printf("✅ %s\n", successMsg)
	} else {
		type errResp struct{ Error string }
		var e errResp
		_ = json.NewDecoder(resp.Body).Decode(&e)
		if e.Error != "" {
			fmt.Printf("❌ Ошибка сервера (%d): %s\n", resp.StatusCode, e.Error)
		} else {
			fmt.Printf("❌ Сервер вернул ошибку со статусом: %d\n", resp.StatusCode)
		}
	}
}
