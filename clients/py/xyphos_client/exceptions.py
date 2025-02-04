"""🚨 Xyphos Client Exceptions

This module contains all the custom exceptions used by the Xyphos client.
Each exception is properly documented and includes an emoji for better readability.
"""

class XyphosError(Exception):
    """🚨 Base exception for all Xyphos errors"""
    pass

class AuthenticationError(XyphosError):
    """🔑 Raised when authentication fails (e.g., invalid credentials, expired token)"""
    pass

class NotFoundError(XyphosError):
    """🔍 Raised when a requested resource is not found"""
    pass

class PermissionError(XyphosError):
    """🚫 Raised when the client lacks permission to perform an operation"""
    pass

class InvalidInputError(XyphosError):
    """⚠️ Raised when invalid input is provided to an operation"""
    pass

class RateLimitError(XyphosError):
    """⏰ Raised when API rate limits are exceeded"""
    pass

class ServerError(XyphosError):
    """🔥 Raised when the server encounters an internal error"""
    pass

class NetworkError(XyphosError):
    """🌐 Raised when network-related issues occur"""
    pass

class ConfigurationError(XyphosError):
    """⚙️ Raised when there are client configuration issues"""
    pass

class ValidationError(XyphosError):
    """✅ Raised when data validation fails"""
    pass 