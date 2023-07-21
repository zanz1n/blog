import { Err, Ok, type Result } from "ts-results";
import { JsonUser, User } from "./User";
import { ApiError } from "./errors";

export interface JsonPost {
  id: string;
  createdAt: string;
  updatedAt: string;
  title: string;
  content: string;
  thumbImage: string | null;
  userId: string | null;
  user: JsonUser | null;
}

export function isValidApiPost(payload: unknown): payload is JsonPost {
  return true;
}

export interface PartialPost {
  userId: string;
}

export interface CompletePost {
  user: User;
}

export class Post {
  private constructor(
    public id: string,
    public createdAt: Date,
    public updatedAt: Date,
    public title: string,
    public content: string,
    public thumbImage: string | null,
    public user?: User,
    public userId?: string,
  ) { }

  isPartial(): this is PartialPost {
    if (this.userId) {
      return true;
    }
    return false;
  }

  isComplete(): this is CompletePost {
    if (this.user) {
      return true;
    }
    return false;
  }

  static fromObject(json: unknown): Result<Post, ApiError> {
    if (!isValidApiPost(json)) {
      return Err(ApiError.InvalidPostData);
    }

    const instance = new this(
      json.id,
      new Date(json.createdAt),
      new Date(json.updatedAt),
      json.title,
      json.content,
      json.thumbImage,
    );

    if (json.user) {
      instance.user = User.fromObject(json.user).unwrap();
    } else if (json.userId) {
      instance.userId = json.userId;
    }

    return Ok(instance);
  }

  static fromJson(payload: string): Result<Post, ApiError> {
    let parsed: unknown;

    try {
      parsed = JSON.parse(payload) as unknown;
    } catch (e) {
      return new Err(ApiError.InvalidBodyPayload);
    }

    return this.fromObject(parsed);
  }
}
