package scalr

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

type StorageProfileBackendType string

const (
	StorageProfileBackendTypeGoogle  StorageProfileBackendType = "google"
	StorageProfileBackendTypeAWSS3   StorageProfileBackendType = "aws-s3"
	StorageProfileBackendTypeAzureRM StorageProfileBackendType = "azurerm"
)

var _ StorageProfiles = (*storageProfiles)(nil)

type StorageProfiles interface {
	List(ctx context.Context, options StorageProfileListOptions) (*StorageProfileList, error)
	Create(ctx context.Context, options StorageProfileCreateOptions) (*StorageProfile, error)
	Read(ctx context.Context, spID string) (*StorageProfile, error)
	Update(ctx context.Context, spID string, options StorageProfileUpdateOptions) (*StorageProfile, error)
	Delete(ctx context.Context, spID string) error
}

type StorageProfile struct {
	ID          string                    `jsonapi:"primary,storage-profiles"`
	Name        string                    `jsonapi:"attr,name"`
	Default     bool                      `jsonapi:"attr,default"`
	IsSystem    bool                      `jsonapi:"attr,is-system"`
	BackendType StorageProfileBackendType `jsonapi:"attr,backend-type"`
	// Unfortunately the existing Scalr default profile has all-zero create date
	// that cannot be parsed as ISO8601 or Unix timestamp and it would fail the unmarshalling.
	// Treating it as a string instead.
	CreatedAt    string     `jsonapi:"attr,created-at"`
	UpdatedAt    *time.Time `jsonapi:"attr,updated-at,iso8601"`
	ErrorMessage *string    `jsonapi:"attr,error-message"`

	GoogleCredentials   *json.RawMessage `jsonapi:"attr,google-credentials"`
	GoogleEncryptionKey *string          `jsonapi:"attr,google-encryption-key"`
	GoogleProject       *string          `jsonapi:"attr,google-project"`
	GoogleStorageBucket *string          `jsonapi:"attr,google-storage-bucket"`

	AWSS3Audience   *string `jsonapi:"attr,aws-s3-audience"`
	AWSS3BucketName *string `jsonapi:"attr,aws-s3-bucket-name"`
	AWSS3Region     *string `jsonapi:"attr,aws-s3-region"`
	AWSS3RoleArn    *string `jsonapi:"attr,aws-s3-role-arn"`

	AzureRMAudience       *string `jsonapi:"attr,azurerm-audience"`
	AzureRMClientID       *string `jsonapi:"attr,azurerm-client-id"`
	AzureRMContainerName  *string `jsonapi:"attr,azurerm-container-name"`
	AzureRMStorageAccount *string `jsonapi:"attr,azurerm-storage-account"`
	AzureRMTenantID       *string `jsonapi:"attr,azurerm-tenant-id"`
}

type StorageProfileList struct {
	*Pagination
	Items []*StorageProfile
}

type StorageProfileListOptions struct {
	ListOptions
	Query  *string                   `url:"query,omitempty"`
	Filter *StorageProfileListFilter `url:"filter,omitempty"`
}

type StorageProfileListFilter struct {
	ID      *string `url:"storage-profile,omitempty"`
	Name    *string `url:"name,omitempty"`
	Default *bool   `url:"default,omitempty"`
}

type StorageProfileCreateOptions struct {
	ID string `jsonapi:"primary,storage-profiles"`

	Name        string                    `jsonapi:"attr,name"`
	Default     *bool                     `jsonapi:"attr,default,omitempty"`
	BackendType StorageProfileBackendType `jsonapi:"attr,backend-type"`

	GoogleCredentials   *json.RawMessage `jsonapi:"attr,google-credentials,omitempty"`
	GoogleEncryptionKey *string          `jsonapi:"attr,google-encryption-key,omitempty"`
	GoogleProject       *string          `jsonapi:"attr,google-project,omitempty"`
	GoogleStorageBucket *string          `jsonapi:"attr,google-storage-bucket,omitempty"`

	AWSS3Audience   *string `jsonapi:"attr,aws-s3-audience,omitempty"`
	AWSS3BucketName *string `jsonapi:"attr,aws-s3-bucket-name,omitempty"`
	AWSS3Region     *string `jsonapi:"attr,aws-s3-region,omitempty"`
	AWSS3RoleArn    *string `jsonapi:"attr,aws-s3-role-arn,omitempty"`

	AzureRMAudience       *string `jsonapi:"attr,azurerm-audience,omitempty"`
	AzureRMClientID       *string `jsonapi:"attr,azurerm-client-id,omitempty"`
	AzureRMContainerName  *string `jsonapi:"attr,azurerm-container-name,omitempty"`
	AzureRMStorageAccount *string `jsonapi:"attr,azurerm-storage-account,omitempty"`
	AzureRMTenantID       *string `jsonapi:"attr,azurerm-tenant-id,omitempty"`
}

type StorageProfileUpdateOptions struct {
	ID string `jsonapi:"primary,storage-profiles"`

	Name        *string                    `jsonapi:"attr,name,omitempty"`
	Default     *bool                      `jsonapi:"attr,default,omitempty"`
	BackendType *StorageProfileBackendType `jsonapi:"attr,backend-type,omitempty"`

	GoogleCredentials   *json.RawMessage `jsonapi:"attr,google-credentials,omitempty"`
	GoogleEncryptionKey *string          `jsonapi:"attr,google-encryption-key,omitempty"`
	GoogleProject       *string          `jsonapi:"attr,google-project,omitempty"`
	GoogleStorageBucket *string          `jsonapi:"attr,google-storage-bucket,omitempty"`

	AWSS3Audience   *string `jsonapi:"attr,aws-s3-audience,omitempty"`
	AWSS3BucketName *string `jsonapi:"attr,aws-s3-bucket-name,omitempty"`
	AWSS3Region     *string `jsonapi:"attr,aws-s3-region,omitempty"`
	AWSS3RoleArn    *string `jsonapi:"attr,aws-s3-role-arn,omitempty"`

	AzureRMAudience       *string `jsonapi:"attr,azurerm-audience,omitempty"`
	AzureRMClientID       *string `jsonapi:"attr,azurerm-client-id,omitempty"`
	AzureRMContainerName  *string `jsonapi:"attr,azurerm-container-name,omitempty"`
	AzureRMStorageAccount *string `jsonapi:"attr,azurerm-storage-account,omitempty"`
	AzureRMTenantID       *string `jsonapi:"attr,azurerm-tenant-id,omitempty"`
}

type storageProfiles struct {
	client *Client
}

func (s *storageProfiles) List(ctx context.Context, options StorageProfileListOptions) (*StorageProfileList, error) {
	req, err := s.client.newRequest("GET", "storage-profiles", &options)
	if err != nil {
		return nil, err
	}

	spl := &StorageProfileList{}
	err = s.client.do(ctx, req, spl)
	if err != nil {
		return nil, err
	}

	return spl, nil
}

func (s *storageProfiles) Read(ctx context.Context, spID string) (*StorageProfile, error) {
	u := fmt.Sprintf("storage-profiles/%s", url.QueryEscape(spID))
	req, err := s.client.newRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	sp := &StorageProfile{}
	err = s.client.do(ctx, req, sp)
	if err != nil {
		return nil, err
	}

	return sp, nil
}

func (s *storageProfiles) Create(ctx context.Context, options StorageProfileCreateOptions) (*StorageProfile, error) {
	req, err := s.client.newRequest("POST", "storage-profiles", &options)
	if err != nil {
		return nil, err
	}

	sp := &StorageProfile{}
	err = s.client.do(ctx, req, sp)
	if err != nil {
		return nil, err
	}

	return sp, nil
}

func (s *storageProfiles) Update(ctx context.Context, spID string, options StorageProfileUpdateOptions) (*StorageProfile, error) {
	u := fmt.Sprintf("storage-profiles/%s", url.QueryEscape(spID))
	req, err := s.client.newRequest("PATCH", u, &options)
	if err != nil {
		return nil, err
	}

	sp := &StorageProfile{}
	err = s.client.do(ctx, req, sp)
	if err != nil {
		return nil, err
	}

	return sp, nil
}

func (s *storageProfiles) Delete(ctx context.Context, spID string) error {
	u := fmt.Sprintf("storage-profiles/%s", url.QueryEscape(spID))
	req, err := s.client.newRequest("DELETE", u, nil)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}
