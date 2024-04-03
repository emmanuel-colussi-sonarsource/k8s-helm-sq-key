package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"strings"

	"fmt"
	"os"
	"path/filepath"

	"github.com/golang/glog"
	_ "github.com/lib/pq"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Configuration struct {
	ClusterName  string
	NSDataBase   string
	PvcDBsize    string
	PGSecret     string
	StorageClass string
	Sonaruser    string
	Sonarpass    string
	PGsql        string
	PGconf       string
	PGsvc        string
	NSSonar      string
}

const (
	gcmTagLength = 128
	gcmIVLength  = 12
	cryptoAlgo   = "AES/GCM/NoPadding"
	aesGCMHeader = "{aes-gcm}"
)

type AesGCMCipher struct {
	key []byte
}

func GetConfig(configjs Configuration) Configuration {

	fconfig, err := os.ReadFile("../db/config.json")
	if err != nil {
		panic("❌ Problem with the configuration file : config.json")
		os.Exit(1)
	}
	if err := json.Unmarshal(fconfig, &configjs); err != nil {
		fmt.Println("❌ Error unmarshaling JSON:", err)
		os.Exit(1)
	}

	return configjs
}

func NewAesGCMCipher() *AesGCMCipher {
	// Generate a random key
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		panic(err)
	}
	return &AesGCMCipher{key: key}
}

func (c *AesGCMCipher) Encrypt(clearText string) (string, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", err
	}

	// Generate a random IV
	iv := make([]byte, gcmIVLength)
	if _, err := rand.Read(iv); err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	cipherText := aesgcm.Seal(nil, iv, []byte(clearText), nil)
	encryptedText := append(iv, cipherText...)

	// Base64 encode the encrypted text
	return aesGCMHeader + base64.StdEncoding.EncodeToString(encryptedText), nil
}

func (c *AesGCMCipher) Decrypt(encryptedText string) (string, error) {
	if !strings.HasPrefix(encryptedText, aesGCMHeader) {
		return "", fmt.Errorf("invalid encrypted text format")
	}

	encryptedText = strings.TrimPrefix(encryptedText, aesGCMHeader)

	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", err
	}

	// Decode the Base64 encoded encrypted text
	encryptedData, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", err
	}

	if len(encryptedData) < gcmIVLength {
		return "", fmt.Errorf("invalid encrypted text")
	}

	iv := encryptedData[:gcmIVLength]
	cipherText := encryptedData[gcmIVLength:]

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	plainText, err := aesgcm.Open(nil, iv, cipherText, nil)
	if err != nil {
		return "", err
	}

	return string(plainText), nil
}

func namespaceExists(clientset *kubernetes.Clientset, namespace string) bool {
	_, err := clientset.CoreV1().Namespaces().Get(context.Background(), namespace, metav1.GetOptions{})
	if err != nil {
		return false
	}
	return true
}

func main() {

	// Read config.json file - Get DB Password
	var config1 Configuration
	var AppConfig = GetConfig(config1)

	// Parse command-line arguments
	cmdArgs := os.Args[1:]

	// Load Kubeconfig
	kubeconfigPath := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	config, err := rest.InClusterConfig()
	if err != nil {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		glog.Fatalf("❌ Failed to create a ClientSet: %v. Exiting.", err)
	}

	if len(cmdArgs) != 1 || (cmdArgs[0] != "deploy" && cmdArgs[0] != "destroy") {
		fmt.Println("❌ Usage: go run main.go [deploy|destroy]")
		os.Exit(1)
	}

	/*------------------------- Main -----------------------------*/

	if cmdArgs[0] == "deploy" {

		// Generate AES-GCM key and IV
		cipher := NewAesGCMCipher()

		fmt.Printf("\r%s \n", "Creating namespace...")

		// Check if the namespace exists before creating it
		if !namespaceExists(clientset, AppConfig.NSSonar) {
			nsName := &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: AppConfig.NSSonar,
				},
			}
			_, err = clientset.CoreV1().Namespaces().Create(context.Background(), nsName, metav1.CreateOptions{})
			if err != nil {
				fmt.Printf("❌ Error creating namespace: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("\r✅ Namespace %s created successfully\n", AppConfig.NSSonar)
		} else {
			fmt.Printf("\r✅ Namespace %s exist\n", AppConfig.NSSonar)
		}

		// Create Kubernetes secret for key
		keySecret := &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "sonarsecretkey",
				Namespace: AppConfig.NSSonar,
			},
			Data: map[string][]byte{
				"sonar-secret.txt": []byte(base64.StdEncoding.EncodeToString(cipher.key)),
			},
		}

		// Store key secret in Kubernetes
		_, err = clientset.CoreV1().Secrets(AppConfig.NSSonar).Create(context.Background(), keySecret, metav1.CreateOptions{})
		if err != nil {
			fmt.Println("❌ Error creating Kubernetes key secret:", err)
			os.Exit(1)
		}

		fmt.Println("✅ AES-GCM key stored in Kubernetes secret successfully.")

		// Encrypt the password
		password, err := cipher.Encrypt(AppConfig.Sonarpass)
		if err != nil {
			fmt.Println("❌ Error encrypting password:", err)
			os.Exit(1)
		}

		// Create Kubernetes secret for encrypted password
		passwordSecret := &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "sonarjdbcpassword",
				Namespace: AppConfig.NSSonar,
			},
			Data: map[string][]byte{
				"SONAR_JDBC_PASSWORD": []byte(password),
			},
		}

		// Store password secret in Kubernetes
		_, err = clientset.CoreV1().Secrets(AppConfig.NSSonar).Create(context.Background(), passwordSecret, metav1.CreateOptions{})
		if err != nil {
			fmt.Println("❌ Error creating Kubernetes password secret:", err)
			os.Exit(1)
		}

		fmt.Println("✅ Encrypted password stored in Kubernetes secret successfully.")

		// Test decrypt
		/*	decryptedPassword, err := cipher.Decrypt(password)
			if err != nil {
				fmt.Printf("❌ Error decrypting password: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("✅ Decrypted Password: %s\n", decryptedPassword)*/

	} else if cmdArgs[0] == "destroy" {

		/*--------------------------------- Destroy Steps ------------------------------------*/

		// Delete the secret : sonarjdbcpassword
		err = clientset.CoreV1().Secrets(AppConfig.NSSonar).Delete(context.TODO(), "sonarjdbcpassword", metav1.DeleteOptions{})
		if err != nil {
			fmt.Printf("\n❌ Error deleting secret %s: %v\n", "sonarjdbcpassword", err)
			os.Exit(1)
		}

		fmt.Println("\n✅ Secret %s deleted successfully :", "sonarjdbcpassword")

		// Delete the secret : ssonarsecretkey
		err = clientset.CoreV1().Secrets(AppConfig.NSSonar).Delete(context.TODO(), "sonarsecretkey", metav1.DeleteOptions{})
		if err != nil {
			fmt.Printf("\n❌ Error deleting secret %s: %v\n", "sonarsecretkey", err)
			os.Exit(1)
		}

		fmt.Println("\n✅ Secret %s deleted successfully :", "sonarsecretkey")

	}
}
