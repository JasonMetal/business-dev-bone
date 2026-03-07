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

// core: cdn-eos errors.
const (
	// ErrCodeNoSuchKey - 404: Entry not found.
	ErrCodeNoSuchKey int = iota + 110300
)

// core: project errors.
const (
	// ErrProjectNotFound - 404: Project not found.
	ErrProjectNotFound int = iota + 110350
)

// core: game server profile errors.
const (
	// ErrProjectProfileNotFound - 404: Game server profile not found.
	ErrProjectProfileNotFound int = iota + 110400
	// ErrDelProfileHistoryActivated - 404: 激活状态无法删除.
	ErrDelProfileHistoryActivated
	// ErrProfileHistoryNoTestSuccess - 404: 配置未测试成功无法应用.
	ErrProfileHistoryNoTestSuccess
)

// core: region errors.
const (
	// ErrRegionNotFound - 404: Region not found.
	ErrRegionNotFound int = iota + 110450
)

// core: room errors.
const (
	// ErrRoomNotFound - 404: Room not found.
	ErrRoomNotFound int = iota + 110500
	// ErrRoomStateUpdateForbid - 400: Room only supports Ready and Running state update.
	ErrRoomStateUpdateForbid
)

// core: room player errors.
const (
	// ErrRoomPlayerNotFound - 404: Player not found.
	ErrRoomPlayerNotFound int = iota + 110550
)
const (
	// ErrCreateTickets - 500: Tickets create err.
	ErrCreateTickets int = iota + 110650
	// ErrCrateBackFill - 500: Inst create err.
	ErrCrateBackFill
	// ErrNewAnyInst - 500: Ticket encode err.
	ErrNewAnyInst
	// ErrNotFoundTickets - 404: Tickets not found.
	ErrNotFoundTickets
	// ErrGetBackFill - 500: Get Backfill err.
	ErrGetBackFill
	// ErrNotFoundBackFill - 404: backFill not found.
	ErrNotFoundBackFill
	// ErrStopBackFill - 500: Stop Backfill err.
	ErrStopBackFill
	// ErrUpdateBackFill - 500: Update Backfill err.
	ErrUpdateBackFill
)

const (
	// ErrImageDeleteNotFound - 404: Image not found.
	ErrImageDeleteNotFound int = iota + 110850
)

// core: ris-room-profile errors.
const (
	// ErrRisRoomProfileNotFound - 404: Room profile not found.
	ErrRisRoomProfileNotFound int = iota + 110900
	// ErrRisRoomTierNotFound - 404: Room tier not found.
	ErrRisRoomTierNotFound
)

// core: ris-room errors.
const (
	// ErrRisRoomNotFound - 404: Room not found.
	ErrRisRoomNotFound int = iota + 110920
	// ErrExceedingMaxRoomLimit - 400: Exceeding the maximum room limit.
	ErrExceedingMaxRoomLimit
	// ErrExceedingMaxPlayerLimit - 400: Exceeding the maximum player limit.
	ErrExceedingMaxPlayerLimit
	// ErrRisRoomNotOpen - 400: Room is not open.
	ErrRisRoomNotOpen
	// ErrRisRoomAlreadyExist - 400: RisRoom already exist.
	ErrRisRoomAlreadyExist
	// ErrRisRoomUniqueNameEnabled - 400: Please reconfigure the switch.
	ErrRisRoomUniqueNameEnabled
	// ErrRisRoomCreateAllocSrvFailed - 400: Create Allocation service failed.
	ErrRisRoomCreateAllocSrvFailed
)

// core: ris-plugin errors.
const (
	// ErrRisPluginNotFound - 404: Plugin not found.
	ErrRisPluginNotFound int = iota + 110960
	// ErrRisPluginReturnTypeCancel - 400: Plugin return type cancel.
	ErrRisPluginReturnTypeCancel
)

// project_basic_setup
const (
	// ErrProjectBasicSetupNotFound - 404: Project basic setup not found.
	ErrProjectBasicSetupNotFound int = iota + 111000
)

// rcs_config
const (
	// ErrRcsConfigNotFound - 404: Rcs config not found.
	ErrRcsConfigNotFound int = iota + 111100
	// ErrRcsConfigDuplicated - 404: The same keyName cannot exist in the same MOS application.
	ErrRcsConfigDuplicated
	// ErrRcsConfigCannotExportAsExcel - 404: The current configuration data cannot be exported as a spreadsheet. 当前配置数据不能以电子表格形式导出。
	ErrRcsConfigCannotExportAsExcel
	// ErrRcsConfigCannotDeleteExcelSheet1 - 404: Default Sheet1 cannot be deleted. 无法删除默认Sheet1。
	ErrRcsConfigCannotDeleteExcelSheet1
	// ErrRcsVersionNotFound - 404: RcsVersion not found.
	ErrRcsVersionNotFound
	// ErrRcsVersionDuplicated - 404: The same RcsVersion cannot exist in the same MOS application.版本重复。
	ErrRcsVersionDuplicated
)

// cas_archive
const (
	// ErrCasArchiveNotFound - 404:  Cas archive not found.
	ErrCasArchiveNotFound int = iota + 111200
	// ErrCasArchiveAlreadyExists - 404:  Archives already exist and cannot be created repeatedly. 存档已存在，不可重复创建。
	ErrCasArchiveAlreadyExists
	// ErrCasArchiveArchiveContentKeyOrValueIsEmpty - 404:  Cas archive ArchiveContent or ArchiveContentKey or ArchiveContentValue is empty.
	ErrCasArchiveArchiveContentKeyOrValueIsEmpty
	// ErrCasArchiveUnmarshalContentException - 500:  Cas archive unmarshal content exception.
	ErrCasArchiveUnmarshalContentException
	// ErrCasArchiveMarshalContentException - 500:  Cas archive marshal content exception.
	ErrCasArchiveMarshalContentException

	// ErrCasUnsupportedArchiveContentKeyException - 500:  Unsupported type for archiveContentKey exception.
	ErrCasUnsupportedArchiveContentKeyException
)

// t_daily_traffic_stat
const (
	// ErrDailyTrafficStatDuplicated - 404: The same MOS application cannot count the traffic data of the same date twice.
	ErrDailyTrafficStatDuplicated int = iota + 111300
)

const (
	// ErrNotAutoScale - 404: auto scale close.
	ErrNotAutoScale int = iota + 111400
	// ErrNotRoomProfile - 404: 暂无对应规模的房间配置.
	ErrNotRoomProfile
)

// k8s
const (
	// ErrK8sPodNotFound - 404: Pod not found.
	ErrK8sPodNotFound int = iota + 111500
	// ErrK8sForbidden - 403: Forbidden.
	ErrK8sForbidden
	// ErrK8sInvalidParam - 400: Invalid parameter.
	ErrK8sInvalidParam
	// ErrK8sInternalError - 500: Internal error.
	ErrK8sInternalError
)

// agones
const (
	// ErrAgonesGameServerNotFound - 404: GameServer not found.
	ErrAgonesGameServerNotFound int = iota + 111600
	// ErrAgonesForbidden - 403: Forbidden.
	ErrAgonesForbidden
	// ErrAgonesInvalidParam - 400: Invalid parameter.
	ErrAgonesInvalidParam
	// ErrAgonesInternalError - 500: Internal error.
	ErrAgonesInternalError
	// ErrAgonesFleetUnAllocated - 400: 舰队无法分配.
	ErrAgonesFleetUnAllocated
)

// passport.realm
const (
	// ErrRealmNotFound - 404: Realm not found.
	ErrRealmNotFound int = iota + 111700
	// ErrRealmAlreadyExist - 400: Realm already exist.
	ErrRealmAlreadyExist
)

const (
	// ErrUpdateScoreDataNotInTime - 500: 排行榜不在有效期内无法更新数据.
	ErrUpdateScoreDataNotInTime int = iota + 111800
	// ErrLeaderboardConfigSortFiled - 500: 排行榜排序字段配置有误.
	ErrLeaderboardConfigSortFiled
	// ErrSetLeaderboardRestTime - 500: 排行榜下次更新时间设置有误.
	ErrSetLeaderboardRestTime
	// ErrGetLeaderboardRecord - 500: 获取数据异常.
	ErrGetLeaderboardRecord
	// ErrLeaderboardReset - 500: 设置重置失败.
	ErrLeaderboardReset
)

// friend_relation
const (
	// ErrFriendRelationNotFound - 404: Friend Relation not found.你和你当前的角色id的好友关系不存在！
	ErrFriendRelationNotFound int = iota + 111900
	// ErrFriendRelationDuplicated - 404: The same appId friend relation cannot exist in the same MOS application.
	ErrFriendRelationDuplicated
	// ErrFriendNumLimit - 404: Friend Num is limit.超过拥有好友的数量限制。
	ErrFriendNumLimit
	// ErrFriendNumLimitFromPersonaID  - 404: 自己好友列表已满。
	ErrFriendNumLimitFromPersonaID
	// ErrFriendNumLimitTargetPersonaID  - 404: 对方好友列表已满。
	ErrFriendNumLimitTargetPersonaID
	// ErrCannotBlockYourself - 404: You can't block yourself.自己不能拉黑自己.
	ErrCannotBlockYourself
)

// friend_blacklist
const (
	// ErrFriendBlacklistNotFound - 404: FriendBlacklist not found.
	ErrFriendBlacklistNotFound int = iota + 112000
	// ErrFriendBlacklistDuplicated - 404:  FriendBlacklist data is duplicated.您已拉黑对方。
	ErrFriendBlacklistDuplicated
)

// t_user_info
const (
	// ErrUserInfoNotFound - 404: UserInfo not found.
	ErrUserInfoNotFound int = iota + 112100
)

// t_friend_presence
const (
	// ErrFriendPresenceNotFound - 404: FriendPresence not found.
	ErrFriendPresenceNotFound int = iota + 112200
	// ErrFriendPresenceDuplicated - 404: FriendPresence data is duplicated.
	ErrFriendPresenceDuplicated
	// ErrFriendPresenceUnmarshalContentException - 500: FriendPresence unmarshal content exception.
	ErrFriendPresenceMarshalContentException
	// ErrInvalidParameter - 500: ErrInvalidParameter.
	ErrInvalidParameter
	// ErrRedisSet - 500: ErrRedisSet.
	ErrRedisSet
)

// friend_config
const (
	// ErrFriendConfigNotFound - 404: Friend config not found.
	ErrFriendConfigNotFound int = iota + 112300
	// ErrFriendConfigDuplicated - 404: The same appId friend config cannot exist in the same MOS application.
	ErrFriendConfigDuplicated
)

// friend_request
const (
	// ErrFriendRequestNotFound - 404: Friend Request not found.
	ErrFriendRequestNotFound int = iota + 112400
	// ErrFriendRequestLimit - 404: Friend Request is limit.超过了好友请求上限数。
	ErrFriendRequestLimit
	// ErrFriendRequestDuplicated - 404: The same appId friend request cannot exist in the same MOS application. 同一个 appId 好友请求不能存在于同一个 MOS 应用程序中。
	ErrFriendRequestDuplicated
	// ErrFriendRequestAlreadyIsPending - 404:  You have applied for a friend request, pending approval.
	ErrFriendRequestAlreadyIsPending
	// ErrFriendRelationAlreadyIsApproved - 404:  You are already friends.
	ErrFriendRelationAlreadyIsApproved
	// ErrFriendRelationIsBlocked - 404:  You are blocked.
	ErrFriendRelationIsBlocked
	// ErrFriendRequestDuplicatedFriendRequestIsApproved - 404: The same appId friend request is Approved and cannot exist in the same MOS application.当前角色双方已经是好友关系，不需要重复申请（拒绝或通过等其他操作）。
	ErrFriendRequestDuplicatedFriendRequestIsApproved
	// ErrNotObtainTheRoleIDAssignedByTheMosSystem - 404: If the PP login fails or is abnormal, you do not obtain the role ID assigned by the MOS system.
	ErrNotObtainTheRoleIDAssignedByTheMosSystem
	// ErrFriendRequestCanceled - 404: The other party has withdrawn the friend request.对方已撤销申请。
	ErrFriendRequestCanceled
)

const (
	// ErrVerifySMSCode - 401: 短信验证失败.
	ErrVerifySMSCode int = iota + 112500
	// ErrNotRefreshToken - 401: 缺少Refresh Token.
	ErrNotRefreshToken
	// ErrExpiredToken - 401: Token已失效.
	ErrExpiredToken
	// ErrDeviceFingerprintVerify - 401: 设备验证失败.
	ErrDeviceFingerprintVerify
)

// role
const (
	// ErrFriendRoleNotFound - 404: Role not found . 角色不存在，或已被管理员删除。
	ErrFriendRoleNotFound int = iota + 112600
	// ErrMissingPersonaId - 401: The `PersonaId` was empty. 角色id不存在。
	ErrMissingPersonaId
	// ErrCurrentNotAllowMultiRoles - 404: The current domain does not allow the creation of multiple roles.
	ErrCurrentNotAllowMultiRoles
	// ErrFriendRoleDuplicated - 404: The same Role cannot exist in the same MOS application.
	ErrFriendRoleDuplicated
	// ErrFriendRoleUnmarshalContentException - 500: FriendRole unmarshal content exception.
	ErrFriendRoleUnmarshalContentException
	// ErrFriendRoleIdsLengthLimit - 404: Ids length must be less than 50. 角色ids里的id个数不能超过50个。
	ErrFriendRoleIdsLengthLimit
	// ErrMissingFriendRoleIds - 404: 角色ids不存在。
	ErrMissingFriendRoleIds
)

// role domain config
const (
	// ErrFriendRoleDomainConfigNotFound - 404:  The friend configuration does not exist. 好友配置不存在。
	ErrFriendRoleDomainConfigNotFound int = iota + 112700
	// ErrFriendRoleDomainConfigDuplicated - 404: Under the same appId, the friend configuration is duplicated. 在同个appId下，好友配置重复。
	ErrFriendRoleDomainConfigDuplicated
)

// PushChannel
const (
	// ErrPushChannelNotFound - 404: PushChannel not found.
	ErrPushChannelNotFound int = iota + 112800
	// ErrPushChannelDuplicated - 404: The current channel name is repeated, cannot allowed in the same MOS application.
	ErrPushChannelDuplicated
	// ErrPushChannelPublicMaxNum - 404: The number of public channels is exceeded.超过公共频道限制数量。
	ErrPushChannelPublicMaxNum
	// ErrPushChannelNonPublicMaxNum - 404:The number of non-public channels is exceeded.超过非公共频道限制数量。
	ErrPushChannelNonPublicMaxNum
)

// PushChannelMember
const (
	// ErrPushChannelMemberNotFound - 404: PushChannelMember not found.
	ErrPushChannelMemberNotFound int = iota + 112900
	// ErrPushChannelMemberDuplicated - 404: The current channel name is repeated, cannot allowed in the same MOS application.
	ErrPushChannelMemberDuplicated
	// ErrPushChannelNotOwner - 404: If you're not the owner of Current Channel, you can't delete it.
	ErrPushChannelNotOwner
	// ErrPushChannelMemberPublicMaxNum - 404: The maximum number of public channels that can be created is 5000
	ErrPushChannelMemberPublicMaxNum
	// ErrPushChannelMemberNonPublicMaxNum - 404: The maximum number of non-public channels that can be created is 500
	ErrPushChannelMemberNonPublicMaxNum
	// ErrPushChannelMemberPrivateMaxNum - 404: The maximum number of private channels that can be created is 2
	ErrPushChannelMemberPrivateMaxNum

	// ErrPushChannelRecentContactNotFound - 404: PushChannelRecentContact not found.
	ErrPushChannelRecentContactNotFound
	// ErrPushChannelRecentContactDuplicated - 404: The recent contact list does not allow duplicate senders and receivers. 最近联系人列表不允许重复同一个发送者和接收者.
	ErrPushChannelRecentContactDuplicated
)

// PushMessage
const (
	// ErrPushChannelMessageNotFound - 404: PushMessage not found.
	ErrPushChannelMessageNotFound int = iota + 113000
	// ErrPushChannelMessageDuplicated - 404: The current channel name is repeated, cannot allowed in the same MOS application.
	ErrPushChannelMessagelDuplicated
	// ErrRateLimitExceeded - 404: The frequency of sending messages exceeds the limit.消息发送频率超过限制.
	ErrRateLimitExceeded
)

const (
	// ErrBillboardNotFound - 404: ErrBillboardNotFound.
	ErrBillboardNotFound int = iota + 114000
	// ErrBillboardDuplicated - 404: The same billboard cannot exist in the same MOS application.
	ErrBillboardDuplicated
	// ErrDataInvalidParam - 400: Invalid parameter.
	ErrDataInvalidParam
)

const (
	// ErrPropXmlData - 200: 数据校验异常.
	ErrPropXmlData int = iota + 113500
	// ErrReadFile - 200: 读取文件异常.
	ErrReadFile
	// ErrPropReadXml - 200: 读取Execl文件异常.
	ErrPropReadXml
	// ErrFileSize - 403: 文件大小超过限制.
	ErrFileSize
	// ErrFileExt - 403: 文件格式为xls或者xlsx.
	ErrFileExt
	// ErrPropVersionNoDuplicated - 403: 版本号重复.
	ErrPropVersionNoDuplicated
	// ErrPropIdDuplicated - 403: 道具唯一标识重复.
	ErrPropIdDuplicated
	// ErrPropVersionNotFound - 404: 版本不存在.
	ErrPropVersionNotFound
)
const (
	// ErrBackpackLackOfProp  - 200: 背包道具不足.
	ErrBackpackLackOfProp int = iota + 113600
	// ErrBackpackBusy  - 200: 服务繁忙，请重试.
	ErrBackpackBusy
	// ErrBackpackPropExceedLimit  - 200: 道具数量超过上限.
	ErrBackpackPropExceedLimit
	// ErrNotFoundPropID  - 200: 未找到道具数据.
	ErrNotFoundPropID
	// ErrBackpackInstIdInvalid  - 200: 背包资源ID无效.
	ErrBackpackInstIdInvalid
)
const (
	// ErrShopPurchaseCost  - 200: 道具不足，购买失败.
	ErrShopPurchaseCost int = iota + 113700
	// ErrShopPurchaseNotCount  - 200: 购买次数不足.
	ErrShopPurchaseNotCount
	// ErrShopVersionNoDuplicated - 403: 版本号重复.
	ErrShopVersionNoDuplicated
	// ErrShopIdNoExist - 200: 商店id不存在.
	ErrShopIdNoExist
	// ErrShopNotEffective - 200: 商店不在有效期.
	ErrShopNotEffective
	// ErrShopItemNoExist - 200: 商品id不存在.
	ErrShopItemNoExist
)
const (
	// ErrInboxNotFound - 404: ErrInboxNotFound.
	ErrInboxNotFound int = iota + 113200
	// ErrInboxDuplicated - 404: The same Inbox cannot exist in the same MOS application.
	ErrInboxDuplicated
	// 	ErrDataUnmarshalContentException - 404: ErrDataUnmarshalContentException.
	ErrDataUnmarshalContentException
	// ErrInvalidParamExpiredTime - 400: The expiration time cannot be earlier than (or equal to) the current time. 过期时间不能早于(或等于)当前时间。
	ErrInvalidParamExpiredTime
	// ErrInvalidParamExpiredTimeAndExpiredHours - 400: The specified time And the specified duration cannot be passed in at the same time. 指定时刻，指定时长不能同时传入。
	ErrInvalidParamExpiredTimeAndExpiredHours
	// ErrInboxAttachmentNotFound - 400: ErrInboxAttachmentNotFound.
	ErrInboxAttachmentNotFound
	// ErrInboxRecordNotFound - 404: ErrInboxRecordNotFound.
	ErrInboxRecordNotFound
	// ErrInboxRecordDuplicated - 404: The same InboxRecord cannot exist in the same MOS application.
	ErrInboxRecordDuplicated
	// ErrInboxMsgNotFound - 404: ErrInboxMsgNotFound.
	ErrInboxMsgNotFound
	// ErrInboxMsgDuplicated - 404: The same InboxMsg cannot exist in the same MOS application.
	ErrInboxMsgDuplicated
	// ErrInvalidParamSendTime - 400: The send time cannot be earlier than (or equal to) the current time. 接收开始时间不能早于(或等于)当前时间。
	ErrInvalidParamSendTime
	// ErrInvalidParamSendExpiredTime - 400: The receive cut-off time cannot be earlier than (or equal to) the current time. 接收截止时间不能早于(或等于)当前时间。
	ErrInvalidParamSendExpiredTime
	// ErrInvalidParamSendExpiredTimeBeEarlierStartTime - 400: The receive cut-off time cannot be earlier than (or equal to) the start time. 接收截止时间不能早于(或等于)开始时间。
	ErrInvalidParamSendExpiredTimeBeEarlierStartTime
	// ErrInvalidParamSendRange - 400: The sending range of the current message does not match. 当前邮件的发送范围不匹配。
	ErrInvalidParamSendRange
	// ErrInvalidParamPersonaIdWithInboxTargetTypeRealm - 400: The PersonaId is not in the realm scope. 当角色id不在域范围内。
	ErrInvalidParamPersonaIdWithInboxTargetTypeRealm
	// ErrInvalidParamPersonaIdWithInboxTargetTypeAll - 400: The PersonaId is not in the appId scope. 当角色id不在appId范围内。
	ErrInvalidParamPersonaIdWithInboxTargetTypeAll
	// ErrInboxMsgAlreadyConsumed - 400: The current attachment has already been received and cannot be collected repeatedly. 当前附件已经被领过，不能重复领取。
	ErrInboxMsgAlreadyConsumed
)

// marketing_minigame
const (
	// ErrMarketingMinigameNotFound - 404: MarketingMinigame not found.
	ErrMarketingMinigameNotFound int = iota + 113300
)
const (
	// ErrQuestNotFound - 404: 任务不存在.
	ErrQuestNotFound int = iota + 113400
	// ErrQuestIdDuplicated - 403: 任务id唯一标识重复.
	ErrQuestIdDuplicated
	// ErrQuestVersionNotFound - 404: 任务版本号不存在.
	ErrQuestVersionNotFound
	// ErrQuestVersionNoDuplicated - 403: 任务版本号重复.
	ErrQuestVersionNoDuplicated
	// ErrQuestVersionNoComboOrderDuplicated - 403: 当前任务下的组合内顺序在同个组合内不能重复.
	ErrQuestVersionNoComboOrderDuplicated
	// ErrQuestCurrentlyInEffectDoesNotExist  - 403: 当前生效的任务列表不存在.
	ErrQuestCurrentlyInEffectDoesNotExist
	// ErrQuestRoleCurrentlyNotFound - 404: 当前角色任务不存在.
	ErrQuestRoleCurrentlyNotFound
	// ErrQuestGroupNotFound - 404: 任务集合不存在.
	ErrQuestGroupNotFound
)
const (
	// ErrDataStatsEventTypeNotFound - 404: DataStatsEventType not found.
	ErrDataStatsEventTypeNotFound int = iota + 113900
)
