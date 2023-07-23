export enum ApiError {
  UserNotFound = 4041,
  InvalidUserIdSize = 4001,
  InternalServerError = 5000,
  InvalidBodyPayload = 4000,
  WeakUserPasswordError = 4002,
  UsernameTooBig = 4003,
  UserEmailTooBig = 4004,
  InvalidUserData = 4005,
  UserAlreadyExists = 4070,
  UserUnauthorized = 4010,
  InvalidAuthToken = 4011,
  ExpiredAuthToken = 4012,
  AuthorizationRequired = 4013,
  SignatureAuthNotSupported = 4006,
  InvalidAuthHeaderFormat = 4014,
  DataMutationDenied = 4016,
  PostNotFound = 4042,
  InvalidPostIdSize = 4007,
  InvalidPostThumbIdSize = 4008,
  InvalidPostSlugSize = 4009,
  InvalidPostData = 40010,
  FailedToGetPostDescription = 4043,
  UserUnderTokenInvalidationPasswordChanged = 40151,
  UserUnderTokenInvalidationPermissionChanged = 40152,
  UserUnderTokenInvalidationTooManyAuthFailures = 40153,
  UserUnderTokenInvalidationUserDeleted = 40154,
  UserUnderTokenInvalidationUserRequest = 40155,
}

export function errorResponse(status: number) {
  return new Response(null, { status });
}
