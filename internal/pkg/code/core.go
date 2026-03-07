package code

//go:generate codegen -type=int

// core: pod errors.
const (
	// ErrPodNotFound - 404: Pod not found.
	ErrPodNotFound int = iota + 110001

	// ErrPodAlreadyExist - 400: Pod already exist.
	ErrPodAlreadyExist

	// ErrPodAllocateFailed - 400: Pod allocate failed.
	ErrPodAllocateFailed
)

// core: service errors.
const (
	// ErrServiceNotFound - 404: Service not found.
	ErrServiceNotFound int = iota + 110101

	// ErrDownloadLog - 400: Download log file err.
	ErrDownloadLog

	// ErrDownloadLogNotFound - 404: Log file not found.
	ErrDownloadLogNotFound

	// ErrImageNotFound - 404: Image not found.
	ErrImageNotFound

	// ErrServiceExceedRegionConfig - 400: Service exceed region config.
	ErrServiceExceedRegionConfig
)

// core: cdn-bucket errors.
const (
	// ErrBucketNotFound - 404: Bucket not found.
	ErrBucketNotFound int = iota + 110200

	// ErrBucketAlreadyExist - 400: Bucket already exist.
	ErrBucketAlreadyExist
)

// core: cdn-entry errors.
const (
	// ErrEntryNotFound - 404: Entry not found.
	ErrEntryNotFound int = iota + 110230
	// ErrEntryPathMaximumOf512Chars - 404: 路径长度超过限制，最大允许512个字符
	ErrEntryPathMaximumOf512Chars
)

// core: cdn-release errors.
const (
	// ErrReleaseNotFound - 404: Release not found.
	ErrReleaseNotFound int = iota + 110235
)

// core: cdn-release-entry errors.
const (
	// ErrReleaseEntryNotFound - 404: Entry not found in release.
	ErrReleaseEntryNotFound int = iota + 110245
)

// core: cdn-badge errors.
const (
	// ErrBadgeNotFound - 404: Badge not found.
	ErrBadgeNotFound int = iota + 110260

	// ErrBadgeAlreadyExist - 400: Badge already exist.
	ErrBadgeAlreadyExist

	// ErrBadgeAlreadyLocked - 400: Badge already locked.
	ErrBadgeAlreadyLocked

	// ErrLatestBadgeProhibitDelete - 400: Latest badge can't be deleted.
	ErrLatestBadgeProhibitDelete

	// ErrLatestBadgeProhibitCreated - 400: Latest badge can't be created.
	ErrLatestBadgeProhibitCreated
)
