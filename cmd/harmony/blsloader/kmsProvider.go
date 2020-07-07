package blsloader

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh/terminal"
)

type AwsConfig struct {
	AccessKey string `json:"aws-access-key-id"`
	SecretKey string `json:"aws-secret-access-key"`
	Region    string `json:"aws-region"`
	Token     string `json:"aws-token,omitempty"`
}

func (cfg *AwsConfig) toAws() *aws.Config {
	cred := credentials.NewStaticCredentials(cfg.AccessKey, cfg.SecretKey, cfg.Token)
	return &aws.Config{
		Region:      aws.String(cfg.Region),
		Credentials: cred,
	}
}

// kmsClientProvider provides the kms client. Implemented by
//   baseKMSProvider - abstract implementation
//   sharedKMSProvider - provide the client with default .aws folder
//   fileKMSProvider - provide the aws config with a json file
//   promptKMSProvider - provide the config field from prompt with time out
type kmsClientProvider interface {
	// getKMSClient returns the KMSClient of the kmsClientProvider with lazy loading.
	getKMSClient() (*kms.KMS, error)

	// getAWSConfig returns the AwsConfig for different implementations
	getAWSConfig() (*AwsConfig, error)

	// toStr return the string presentation of kmsClientProvider
	toStr() string
}

// baseKMSProvider provide the kms client with singleton initialization through
// function getConfig for aws credential and regions loading.
type baseKMSProvider struct {
	client *kms.KMS
	err    error
	once   sync.Once
}

func (provider *baseKMSProvider) getKMSClient() (*kms.KMS, error) {
	provider.once.Do(func() {
		cfg, err := provider.getAWSConfig()
		if err != nil {
			provider.err = err
			return
		}
		provider.client, provider.err = kmsClientWithConfig(cfg)
	})
	if provider.err != nil {
		return nil, provider.err
	}
	return provider.client, nil
}

func (provider *baseKMSProvider) getAWSConfig() (*AwsConfig, error) {
	return nil, errors.New("not implemented")
}

func (provider *baseKMSProvider) toStr() string {
	return "not implemented"
}

// sharedKMSProvider provide the kms session with the default aws config
// locates in directory $HOME/.aws/config
type sharedKMSProvider struct {
	baseKMSProvider
}

func newSharedKMSProvider() *sharedKMSProvider {
	return &sharedKMSProvider{
		baseKMSProvider{},
	}
}

func (provider *sharedKMSProvider) getAWSConfig() (*AwsConfig, error) {
	return nil, nil
}

func (provider *sharedKMSProvider) toStr() string {
	return "shared aws config"
}

// fileKMSProvider provide the kms session from a file with json data of structure
// AwsConfig
type fileKMSProvider struct {
	baseKMSProvider

	file string
}

func newFileKMSProvider(file string) *fileKMSProvider {
	return &fileKMSProvider{
		baseKMSProvider: baseKMSProvider{},
		file:            file,
	}
}

func (provider *fileKMSProvider) getAWSConfig() (*AwsConfig, error) {
	b, err := ioutil.ReadFile(provider.file)
	if err != nil {
		return nil, err
	}
	var cfg *AwsConfig
	if err := json.Unmarshal(b, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (provider *fileKMSProvider) toStr() string {
	return fmt.Sprintf("file %v", provider.file)
}

// promptKMSProvider provide a user interactive console for AWS config.
// Three fields are asked:
//    1. AccessKey  	2. SecretKey		3. Region
// Each field is asked with a timeout mechanism.
type promptKMSProvider struct {
	baseKMSProvider

	timeout time.Duration
}

func newPromptKMSProvider(timeout time.Duration) *promptKMSProvider {
	return &promptKMSProvider{
		baseKMSProvider: baseKMSProvider{},
		timeout:         timeout,
	}
}

func (provider *promptKMSProvider) getAWSConfig() (*AwsConfig, error) {
	fmt.Println("Please provide AWS configurations for KMS encoded BLS keys:")
	accessKey, err := provider.prompt("  AccessKey:")
	if err != nil {
		return nil, fmt.Errorf("cannot get aws access key: %v", err)
	}
	secretKey, err := provider.prompt("  SecretKey:")
	if err != nil {
		return nil, fmt.Errorf("cannot get aws secret key: %v", err)
	}
	region, err := provider.prompt("Region:")
	if err != nil {
		return nil, fmt.Errorf("cannot get aws region: %v", err)
	}
	return &AwsConfig{
		AccessKey: accessKey,
		SecretKey: secretKey,
		Region:    region,
		Token:     "",
	}, nil
}

// prompt prompt the user to input a string for a certain field with timeout.
func (provider *promptKMSProvider) prompt(hint string) (string, error) {
	var (
		res string
		err error

		finished = make(chan struct{})
		timedOut = time.After(provider.timeout)
	)

	go func() {
		res, err = provider.threadedPrompt(hint)
		close(finished)
	}()

	for {
		select {
		case <-finished:
			return res, err
		case <-timedOut:
			return "", errors.New("timed out")
		}
	}
}

func (provider *promptKMSProvider) threadedPrompt(hint string) (string, error) {
	fmt.Print(hint)
	b, err := terminal.ReadPassword(syscall.Stdin)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (provider *promptKMSProvider) toStr() string {
	return "prompt"
}

func kmsClientWithConfig(config *AwsConfig) (*kms.KMS, error) {
	if config == nil {
		return getSharedKMSClient()
	}
	return getKMSClientFromConfig(config)
}

func getSharedKMSClient() (*kms.KMS, error) {
	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create aws session")
	}
	return kms.New(sess), err
}

func getKMSClientFromConfig(config *AwsConfig) (*kms.KMS, error) {
	sess, err := session.NewSession(config.toAws())
	if err != nil {
		return nil, err
	}
	return kms.New(sess), nil
}
