import { Err, Ok, type Result } from "ts-results";
import { ApiError } from "./errors";

export enum UserRole {
  Common = "COMMON",
  Admin = "ADMIN",
  Publisher = "PUBLISHER"
}

export function isValidUserRole(payload: string): payload is UserRole {
  if (
    payload == "COMMON" ||
    payload == "ADMIN" ||
    payload == "PUBLISHER"
  ) return true;
  return false;
}

export interface JsonUser {
  id: string;
  createdAt: string;
  updatedAt: string;
  email: string;
  username: string;
  role: UserRole;
}

export function isValidApiUser(payload: unknown): payload is JsonUser {
  if (!payload || typeof payload != "object") return false;
  if (
    "id" in payload && typeof payload.id == "string" &&
    "createdAt" in payload && typeof payload.createdAt == "string" &&
    "updatedAt" in payload && typeof payload.updatedAt == "string" &&
    "email" in payload && typeof payload.email == "string" &&
    "username" in payload && typeof payload.username == "string" &&
    "role" in payload && typeof payload.role == "string" &&
    isValidUserRole(payload["role"])
  ) return true;
  return false;
}

export class User {
  private constructor(
    public id: string,
    public createdAt: Date,
    public updatedAt: Date,
    public email: string,
    public username: string,
    public role: UserRole
  ) { }

  static fromObject(json: unknown): Result<User, ApiError> {
    if (!isValidApiUser(json)) {
      return new Err(ApiError.InvalidUserData);
    }

    return new Ok(new this(
      json.id,
      new Date(json.createdAt),
      new Date(json.updatedAt),
      json.email,
      json.username,
      json.role,
    ));
  }

  static fromJson(payload: string): Result<User, ApiError> {
    let parsed: unknown;

    try {
      parsed = JSON.parse(payload) as unknown;
    } catch (e) {
      return new Err(ApiError.InvalidBodyPayload);
    }

    return this.fromObject(parsed);
  }
}
