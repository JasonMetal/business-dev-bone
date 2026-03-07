package code

//go:generate codegen -type=int
//go:generate codegen -type=int -doc -output ../../../docs/guide/zh-CN/api/error_code_generated.md

// Common: basic errors.
// Code must start with 1xxxxx.
const (
	// ErrSuccess - 200: OK.
	ErrSuccess int = iota + 100001

	// ErrUnknown - 500: Internal server error.
	ErrUnknown

	// ErrBind - 400: Error occurred while binding the request body to the struct.
	ErrBind

	// ErrValidation - 400: Validation failed.
	ErrValidation

	// ErrTokenInvalid - 401: Token invalid.
	ErrTokenInvalid

	// ErrPageNotFound - 404: Page not found.
	ErrPageNotFound

	// ErrObjectNotFound - 200: record not found.
	ErrObjectNotFound
	// ErrService - 200: 服务异常.
	ErrService
)

// common: database errors.
const (
	// ErrDatabase - 500: Database error.
	ErrDatabase int = iota + 100101
)

// common: authorization and authentication errors.
const (
	// ErrEncrypt - 401: Error occurred while encrypting the user password.
	ErrEncrypt int = iota + 100201

	// ErrSignatureInvalid - 401: Signature is invalid.
	ErrSignatureInvalid

	// ErrExpired - 401: Token expired.
	ErrExpired

	// ErrInvalidAuthHeader - 401: Invalid authorization header.
	ErrInvalidAuthHeader

	// ErrMissingHeader - 401: The `Authorization` header was empty.
	ErrMissingHeader

	// ErrMissingHeaderMgAppid - 401: The `mg-appid` header was empty.
	ErrMissingHeaderMgAppid

	// ErrPasswordIncorrect - 401: Password was incorrect.
	ErrPasswordIncorrect

	// ErrPermissionDenied - 403: Permission denied.
	ErrPermissionDenied
	// ErrInvalidPersonaAppIdCompareHeaderMgAppId - 401: 当前登录角色中的`AppId`和header中的`mg-appid`不一致.
	ErrInvalidPersonaAppIdCompareHeaderMgAppId
	// ErrMissingPersonaAppId - 401: 当前登录角色中的`AppId`为空.
	ErrMissingPersonaAppId
	// ErrMissingPersonaUserId - 401: 当前登录角色中的`UserId`为空.
	ErrMissingPersonaUserId
	// ErrMissingPersonaRealmId- 401: 当前登录角色中的`RealmId`为空.
	ErrMissingPersonaRealmId
	// ErrIncorrectPersonaAppId - 401: 当前登录角色中的`AppId`和当前角色所属的`AppId`不一致.
	ErrIncorrectPersonaAppId
	// ErrIncorrectPersonaUserId - 401: 当前登录角色中的`UserId`和当前角色所属的`UserId`不一致.
	ErrIncorrectPersonaUserId
	// ErrIncorrectPersonaRealmId - 401: 当前登录角色中的`RealmId`和当前角色所属的`RealmId`不一致.
	ErrIncorrectPersonaRealmId
	// ErrIncorrectPickPersonaRealmId - 401: 当前登录角色中的`RealmId`和所选角色的`RealmId`不一致.
	ErrIncorrectPickPersonaRealmId
	// ErrIncorrectPickPersonaId - 401: 当前登录(选角后)角色中的`personaId`和 Params 角色中的`personaId`不一致.
	ErrIncorrectPickPersonaId
)

// common: encode/decode errors.
const (
	// ErrEncodingFailed - 500: Encoding failed due to an error with the data.
	ErrEncodingFailed int = iota + 100301

	// ErrDecodingFailed - 500: Decoding failed due to an error with the data.
	ErrDecodingFailed

	// ErrInvalidJSON - 500: Data is not valid JSON.
	ErrInvalidJSON

	// ErrEncodingJSON - 500: JSON data could not be encoded.
	ErrEncodingJSON

	// ErrDecodingJSON - 500: JSON data could not be decoded.
	ErrDecodingJSON

	// ErrInvalidYaml - 500: Data is not valid Yaml.
	ErrInvalidYaml

	// ErrEncodingYaml - 500: Yaml data could not be encoded.
	ErrEncodingYaml

	// ErrDecodingYaml - 500: Yaml data could not be decoded.
	ErrDecodingYaml
)
