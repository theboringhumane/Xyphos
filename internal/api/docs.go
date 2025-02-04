// Package api Xyphos API
//
//	@title						Xyphos API
//	@version					1.0
//	@description				A secure multi-tenant key management system that provides cryptographic operations and key management capabilities.
//	@termsOfService				https://xyphos.io/terms
//
//	@contact.name				API Support
//	@contact.url				https://xyphos.io/support
//	@contact.email				support@xyphos.io
//
//	@license.name				MIT
//	@license.url				https://github.com/theboringhumane/xyphos/blob/main/LICENSE
//
//	@host						localhost:8080
//	@BasePath					/api
//	@schemes					http https
//
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization

// This file is used to generate Swagger documentation.
// The actual documentation is generated using swag init command.

package api

//	@title			Xyphos API
//	@version		1.0
//	@description	A secure multi-tenant key management system that provides cryptographic operations and key management capabilities.
//	@termsOfService	https://lambda-kms.io/terms

//	@contact.name	API Support
//	@contact.url	https://lambda-kms.io/support
//	@contact.email	support@lambda-kms.io

//	@license.name	MIT
//	@license.url	https://github.com/yourusername/xyphos/blob/main/LICENSE

//	@host		localhost:8080
//	@BasePath	/api
//	@schemes	http https

//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and JWT token.

//	@tag.name			Projects
//	@tag.description	Project management endpoints

//	@tag.name			Locations
//	@tag.description	Location management endpoints

//	@tag.name			KeyRings
//	@tag.description	KeyRing management endpoints

//	@tag.name			CryptoKeys
//	@tag.description	Cryptographic key management and operations

//	@tag.name			ClientConfigs
//	@tag.description	Client configuration management

// Health Check
//	@Summary		Health check endpoint
//	@Description	Get the health status of the API
//	@Tags			Health
//	@Produce		json
//	@Success		200	{object}	HealthResponse
//	@Router			/health [get]

// Project Operations

//	@Summary		Create a new project
//	@Description	Create a new KMS project
//	@Tags			Projects
//	@Accept			json
//	@Produce		json
//	@Param			project	body		CreateProjectRequest	true	"Project details"
//	@Success		200		{object}	Project
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects [post]

//	@Summary		List all projects
//	@Description	Get a list of all KMS projects
//	@Tags			Projects
//	@Produce		json
//	@Success		200	{array}		Project
//	@Failure		401	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects [get]

//	@Summary		Get a project
//	@Description	Get details of a specific project
//	@Tags			Projects
//	@Produce		json
//	@Param			projectId	path		string	true	"Project ID"
//	@Success		200			{object}	Project
//	@Failure		401			{object}	ErrorResponse
//	@Failure		404			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectId} [get]

// Location Operations

//	@Summary		List locations
//	@Description	List all locations in a project
//	@Tags			Locations
//	@Produce		json
//	@Param			projectId	path		string	true	"Project ID"
//	@Success		200			{array}		Location
//	@Failure		401			{object}	ErrorResponse
//	@Failure		404			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectId}/locations [get]

//	@Summary		Get a location
//	@Description	Get details of a specific location
//	@Tags			Locations
//	@Produce		json
//	@Param			projectId	path		string	true	"Project ID"
//	@Param			locationId	path		string	true	"Location ID"
//	@Success		200			{object}	Location
//	@Failure		401			{object}	ErrorResponse
//	@Failure		404			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectId}/locations/{locationId} [get]

// KeyRing Operations

//	@Summary		Create a keyring
//	@Description	Create a new keyring in a location
//	@Tags			KeyRings
//	@Accept			json
//	@Produce		json
//	@Param			projectId	path		string					true	"Project ID"
//	@Param			locationId	path		string					true	"Location ID"
//	@Param			keyring		body		CreateKeyRingRequest	true	"KeyRing details"
//	@Success		200			{object}	KeyRing
//	@Failure		400			{object}	ErrorResponse
//	@Failure		401			{object}	ErrorResponse
//	@Failure		404			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectId}/locations/{locationId}/keyrings [post]

//	@Summary		List keyrings
//	@Description	List all keyrings in a location
//	@Tags			KeyRings
//	@Produce		json
//	@Param			projectId	path		string	true	"Project ID"
//	@Param			locationId	path		string	true	"Location ID"
//	@Success		200			{array}		KeyRing
//	@Failure		401			{object}	ErrorResponse
//	@Failure		404			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectId}/locations/{locationId}/keyrings [get]

//	@Summary		Get a keyring
//	@Description	Get details of a specific keyring
//	@Tags			KeyRings
//	@Produce		json
//	@Param			projectId	path		string	true	"Project ID"
//	@Param			locationId	path		string	true	"Location ID"
//	@Param			keyringId	path		string	true	"KeyRing ID"
//	@Success		200			{object}	KeyRing
//	@Failure		401			{object}	ErrorResponse
//	@Failure		404			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectId}/locations/{locationId}/keyrings/{keyringId} [get]

// CryptoKey Operations

//	@Summary		Create a crypto key
//	@Description	Create a new crypto key in a keyring
//	@Tags			CryptoKeys
//	@Accept			json
//	@Produce		json
//	@Param			projectId	path		string					true	"Project ID"
//	@Param			locationId	path		string					true	"Location ID"
//	@Param			keyringId	path		string					true	"KeyRing ID"
//	@Param			key			body		CreateCryptoKeyRequest	true	"CryptoKey details"
//	@Success		200			{object}	CryptoKey
//	@Failure		400			{object}	ErrorResponse
//	@Failure		401			{object}	ErrorResponse
//	@Failure		404			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectId}/locations/{locationId}/keyrings/{keyringId}/keys [post]

//	@Summary		List crypto keys
//	@Description	List all crypto keys in a keyring
//	@Tags			CryptoKeys
//	@Produce		json
//	@Param			projectId	path		string	true	"Project ID"
//	@Param			locationId	path		string	true	"Location ID"
//	@Param			keyringId	path		string	true	"KeyRing ID"
//	@Success		200			{array}		CryptoKey
//	@Failure		401			{object}	ErrorResponse
//	@Failure		404			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectId}/locations/{locationId}/keyrings/{keyringId}/keys [get]

//	@Summary		Get a crypto key
//	@Description	Get details of a specific crypto key
//	@Tags			CryptoKeys
//	@Produce		json
//	@Param			projectId	path		string	true	"Project ID"
//	@Param			locationId	path		string	true	"Location ID"
//	@Param			keyringId	path		string	true	"KeyRing ID"
//	@Param			keyId		path		string	true	"Key ID"
//	@Success		200			{object}	CryptoKey
//	@Failure		401			{object}	ErrorResponse
//	@Failure		404			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectId}/locations/{locationId}/keyrings/{keyringId}/keys/{keyId} [get]

//	@Summary		Rotate a crypto key
//	@Description	Create a new version of a crypto key
//	@Tags			CryptoKeys
//	@Produce		json
//	@Param			projectId	path		string	true	"Project ID"
//	@Param			locationId	path		string	true	"Location ID"
//	@Param			keyringId	path		string	true	"KeyRing ID"
//	@Param			keyId		path		string	true	"Key ID"
//	@Success		200			{object}	CryptoKey
//	@Failure		401			{object}	ErrorResponse
//	@Failure		404			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectId}/locations/{locationId}/keyrings/{keyringId}/keys/{keyId}/rotate [post]

//	@Summary		Encrypt data
//	@Description	Encrypt data using a crypto key
//	@Tags			CryptoKeys
//	@Accept			json
//	@Produce		json
//	@Param			projectId	path		string			true	"Project ID"
//	@Param			locationId	path		string			true	"Location ID"
//	@Param			keyringId	path		string			true	"KeyRing ID"
//	@Param			keyId		path		string			true	"Key ID"
//	@Param			request		body		EncryptRequest	true	"Data to encrypt"
//	@Success		200			{object}	EncryptResponse
//	@Failure		400			{object}	ErrorResponse
//	@Failure		401			{object}	ErrorResponse
//	@Failure		404			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectId}/locations/{locationId}/keyrings/{keyringId}/keys/{keyId}/encrypt [post]

//	@Summary		Decrypt data
//	@Description	Decrypt data using a crypto key
//	@Tags			CryptoKeys
//	@Accept			json
//	@Produce		json
//	@Param			projectId	path		string			true	"Project ID"
//	@Param			locationId	path		string			true	"Location ID"
//	@Param			keyringId	path		string			true	"KeyRing ID"
//	@Param			keyId		path		string			true	"Key ID"
//	@Param			request		body		DecryptRequest	true	"Data to decrypt"
//	@Success		200			{object}	DecryptResponse
//	@Failure		400			{object}	ErrorResponse
//	@Failure		401			{object}	ErrorResponse
//	@Failure		404			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectId}/locations/{locationId}/keyrings/{keyringId}/keys/{keyId}/decrypt [post]

// Client Configuration Operations

//	@Summary		Create client configuration
//	@Description	Create a new client configuration
//	@Tags			ClientConfigs
//	@Accept			json
//	@Produce		json
//	@Param			config	body		CreateClientConfigRequest	true	"Client configuration details"
//	@Success		200		{object}	ClientConfig
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/client-configs [post]

//	@Summary		List client configurations
//	@Description	List all client configurations
//	@Tags			ClientConfigs
//	@Produce		json
//	@Success		200	{array}		ClientConfig
//	@Failure		401	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/client-configs [get]

//	@Summary		Get client configuration
//	@Description	Get details of a specific client configuration
//	@Tags			ClientConfigs
//	@Produce		json
//	@Param			configId	path		string	true	"Configuration ID"
//	@Success		200			{object}	ClientConfig
//	@Failure		401			{object}	ErrorResponse
//	@Failure		404			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/client-configs/{configId} [get]

//	@Summary		Revoke client configuration
//	@Description	Revoke a client configuration
//	@Tags			ClientConfigs
//	@Produce		json
//	@Param			configId	path		string	true	"Configuration ID"
//	@Success		200			{object}	RevokeResponse
//	@Failure		401			{object}	ErrorResponse
//	@Failure		404			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/client-configs/{configId}/revoke [post]
