 ![SonarQube](../images/sonar.png)![AWS](https://img.shields.io/badge/AWS-%23FF9900.svg?style=for-the-badge&logo=amazon-aws&logoColor=white)![Amazon EKS](https://img.shields.io/static/v1?style=for-the-badge&message=Amazon+EKS&color=222222&logo=Amazon+ECS&logoColor=FF9900&label=)![Static Badge](https://img.shields.io/badge/Go-v1.21-blue:)


# Generate an Encryption AES/GCM Key

 ![Flow pods](images/key.png)

 SonarQube requires encryption keys that adhere to precise specifications to ensure compatibility and security. Specifically, the encryption key must adhere to the following criteria:

- **AES/GCM with No Padding:** Advanced Encryption Standard (AES) in Galois/Counter Mode (GCM) is employed for its strong encryption capabilities and efficient performance. GCM mode operates without padding, ensuring that encrypted data remains compact and efficient.
AES-GCM is an authenticated encryption mode that uses the AES block cipher in counter mode with a polynomial MAC based on Galois field multiplication.

- **Random Initialization Vector (IV) of 12 Bytes:** A random initialization vector adds an additional layer of randomness to the encryption process, enhancing security by ensuring unique ciphertexts even when encrypting the same plaintext multiple times. SonarQube mandates a 12-byte random initialization vector for compatibility and security reasons.

- **Authentication Tags of 128 Bits:** Authentication tags, an integral part of the output of GCM mode, serve as cryptographic checksums to verify the integrity of the encrypted data. SonarQube specifies that the authentication tags should be 128 bits in length, providing a robust mechanism for ensuring data integrity.

- **Base64 Encoding:** To facilitate interoperability and ease of integration with SonarQube deployments, the resulting encryption key is encoded in base64 format. Base64 encoding ensures that the encryption key is represented in a human-readable and platform-independent manner, making it suitable for storage in configuration files, environment variables, or Kubernetes secrets.

> [!CAUTION] 
> It's important to note that while these specifications are current as of the latest SonarQube version, they represent an internal implementation subject to change at any time without notice. As a result, externally generated encryption keys may not guarantee compatibility with future SonarQube versions. Therefore, it's imperative for organizations to stay informed about updates and changes to SonarQube's encryption requirements to maintain the security and compatibility of their deployments.

---

## Prerequisites

Before you get started, you’ll need to have these things:

✅ Previous deployment steps are completed

## What does this task do?

- Create a AES/GCM Key
- Crypt a DB password
- Create a namespace sonarqube
- Create a k8s secret for AES/GCM Key : sonarsecretkey
- Create a k8s secret database for JDBC Password: sonarjdbcpassword

## Useful commands

 * `./gen.sh deploy`      deploy this stack 
 * `./gen.sh destroy`     cleaning up stack


## ✅ Setup Environment

Run the following command to automatically install all the required modules based on the go.mod and go.sum files:

```bash
k8s-helm-sq-key:> cd key
k8s-helm-sq-key:/key> go mod download
``` 

## ✅ Using a Go Program to Generate Encryption Key and Encrypt PostgreSQL Password
In order to automate the process of generating an encryption key and encrypting the PostgreSQL password, we can utilize a Go program. This program will not only generate a key following the specifications outlined earlier but also encrypt the default PostgreSQL password (Bench123) that we initialized in the config.json file located in the db directory during the PostgreSQL deployment.

Generate an Encryption AES/GCM Key with the following command :

```bash
k8s-helm-sq-key:/key> ./gen.sh deploy
✅ Namespace sonarqube created successfully
✅ AES-GCM key stored in Kubernetes secret successfully.
✅ Encrypted password stored in Kubernetes secret successfully.
k8s-helm-sq-key:/key>
```

▶️ Verify if sonarsecretkey secret created :
```bash
k8s-helm-sq-key:/key>kubectl -n sonarqube get secret sonarsecretkey
NAME             TYPE     DATA   AGE
sonarsecretkey   Opaque   2      0m58s
k8s-helm-sq-key:/key>
```

▶️ Verify if sonarjdbcpassword secret created :
```bash
k8s-helm-sq-key:/key>kubectl -n sonarqube get secret sonarjdbcpassword
NAME                TYPE     DATA   AGE
sonarjdbcpassword   Opaque   1      1m19s
k8s-helm-sq-key:/key>
```

## ✅ Now we can deploy SonarQube 
 
-----
<table>
<tr style="border: 0px transparent">
	<td style="border: 0px transparent"> <a href="../db/README.md" title="Database deployment">⬅ Previous</a></td><td style="border: 0px transparent"><a href="../sonarqube/README.md" title="SonarQube deployment">Next ➡</a></td><td style="border: 0px transparent"><a href="../README.md" title="home">🏠</a></td>
</tr>
<tr style="border: 0px transparent">
<td style="border: 0px transparent">Database deployment</td><td style="border: 0px transparent">SonarQube deployment</td><td style="border: 0px transparent"></td>
</tr>

</table>

