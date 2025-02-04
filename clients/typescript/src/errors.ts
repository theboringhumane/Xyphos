/**
 * 🚨 Base error class for KMS errors
 */
export class KMSError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'KMSError';
  }
}

/**
 * 🚨 Authentication error
 */
export class AuthenticationError extends KMSError {
  constructor(message: string = 'Authentication failed') {
    super(message);
    this.name = 'AuthenticationError';
  }
}

/**
 * 🚨 Resource not found error
 */
export class NotFoundError extends KMSError {
  constructor(message: string = 'Resource not found') {
    super(message);
    this.name = 'NotFoundError';
  }
}

/**
 * 🚨 Permission denied error
 */
export class PermissionError extends KMSError {
  constructor(message: string = 'Permission denied') {
    super(message);
    this.name = 'PermissionError';
  }
}

/**
 * 🚨 Invalid input error
 */
export class InvalidInputError extends KMSError {
  constructor(message: string = 'Invalid input provided') {
    super(message);
    this.name = 'InvalidInputError';
  }
} 