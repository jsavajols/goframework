package htmlfiles

import (
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	con "github.com/jsavajols/goframework/const"
)

func GetHtmlFile(filename string) string {
	fileContent, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	fileContentString := string(fileContent)
	return fileContentString
}

func UploadFileToS3(filePath, fileName string) error {
	// Créez une nouvelle session AWS en utilisant les clés d'accès OVH et l'endpoint S3 OVH
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(os.Getenv("OVH_REGION")),   // Exemple de région OVH, ajustez selon votre région
		Endpoint:    aws.String(os.Getenv("OVH_ENDPOINT")), // Ajustez selon votre région et le endpoint OVH
		Credentials: credentials.NewStaticCredentials(os.Getenv("OVH_ACCESS_KEY"), os.Getenv("OVH_SECRET_KEY"), ""),
	})
	if err != nil {
		return err
	}

	// Ouvrez le fichier
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Créez un uploader avec la session
	uploader := s3manager.NewUploader(sess)

	// Téléchargez le fichier
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(os.Getenv("OVH_BUCKET")), // Remplacez par le nom de votre bucket
		Key:    aws.String(fileName),
		Body:   file,
	})
	return err
}

func DeleteFileFromS3(fileName string) error {
	// Créez une nouvelle session AWS en utilisant les clés d'accès OVH et l'endpoint S3 OVH
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(os.Getenv("OVH_REGION")),   // Exemple de région OVH, ajustez selon votre région
		Endpoint:    aws.String(os.Getenv("OVH_ENDPOINT")), // Ajustez selon votre région et le endpoint OVH
		Credentials: credentials.NewStaticCredentials(os.Getenv("OVH_ACCESS_KEY"), os.Getenv("OVH_SECRET_KEY"), ""),
	})
	if err != nil {
		return err
	}

	// Créez une nouvelle instance du service S3
	svc := s3.New(sess)

	// Préparez la demande de suppression
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(os.Getenv("OVH_BUCKET")), // Remplacez par le nom de votre bucket
		Key:    aws.String(fileName),
	}

	// Supprimez le fichier
	_, err = svc.DeleteObject(input)
	if err != nil {
		return err
	}

	// Attendez que le fichier soit supprimé
	return svc.WaitUntilObjectNotExists(&s3.HeadObjectInput{
		Bucket: aws.String(os.Getenv("OVH_BUCKET")),
		Key:    aws.String(fileName),
	})
}

func GeneratePresignedURL(objectKey string) (string, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(os.Getenv("OVH_REGION")),
		Endpoint:    aws.String(os.Getenv("OVH_ENDPOINT")),
		Credentials: credentials.NewStaticCredentials(os.Getenv("OVH_ACCESS_KEY"), os.Getenv("OVH_SECRET_KEY"), ""),
	})
	if err != nil {
		return "", err
	}

	// Création du client S3
	svc := s3.New(sess)

	// Définition des paramètres de la requête pré-signée
	req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(os.Getenv("OVH_BUCKET")),
		Key:    aws.String(objectKey),
	})

	// Génération de l'URL pré-signée
	urlStr, err := req.Presign(con.TIMEOUT_PRESING_URL * time.Minute) // URL valide pour 15 minutes
	if err != nil {
		return "", err
	}

	return urlStr, nil
}

func downloadFileFromS3(bucketName, fileKey string) error {
	// Créez une nouvelle session AWS en utilisant les clés d'accès OVH et l'endpoint S3 OVH
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(os.Getenv("OVH_REGION")),   // Exemple de région OVH, ajustez selon votre région
		Endpoint:    aws.String(os.Getenv("OVH_ENDPOINT")), // Ajustez selon votre région et le endpoint OVH
		Credentials: credentials.NewStaticCredentials(os.Getenv("OVH_ACCESS_KEY"), os.Getenv("OVH_SECRET_KEY"), ""),
	})
	if err != nil {
		return err
	}

	// Créez un nouveau client S3
	s3Client := s3.New(sess)

	// Préparez l'objet de téléchargement
	input := &s3.GetObjectInput{
		Bucket: aws.String(os.Getenv("OVH_BUCKET")),
		Key:    aws.String(fileKey),
	}

	// Téléchargez le fichier
	result, err := s3Client.GetObject(input)
	if err != nil {
		return err
	}
	defer result.Body.Close()

	// Créez un fichier local pour sauvegarder le contenu téléchargé
	file, err := os.Create(fileKey + "-downloaded.png")
	if err != nil {
		return err
	}
	defer file.Close()

	// Copiez le contenu dans le fichier local
	_, err = file.ReadFrom(result.Body)
	return err
}
