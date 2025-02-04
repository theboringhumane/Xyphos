class KMSError(Exception):
    """🚨 Base error class for KMS errors"""
    pass

class AuthenticationError(KMSError):
    """🚨 Authentication error"""
    def __init__(self, message: str = "Authentication failed"):
        super().__init__(message)

class NotFoundError(KMSError):
    """🚨 Resource not found error"""
    def __init__(self, message: str = "Resource not found"):
        super().__init__(message)

class PermissionError(KMSError):
    """🚨 Permission denied error"""
    def __init__(self, message: str = "Permission denied"):
        super().__init__(message)

class InvalidInputError(KMSError):
    """🚨 Invalid input error"""
    def __init__(self, message: str = "Invalid input provided"):
        super().__init__(message)

class RateLimitError(KMSError):
    """🚨 Rate limit error"""
    def __init__(self, message: str = "Rate limit exceeded"):
        super().__init__(message)

class ServerError(KMSError):
    """🚨 Server error"""
    def __init__(self, message: str = "Internal server error"):
        super().__init__(message) 