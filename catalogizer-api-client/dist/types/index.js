"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.ValidationError = exports.NetworkError = exports.AuthenticationError = exports.CatalogizerError = void 0;
// Error types
class CatalogizerError extends Error {
    constructor(message, status, code) {
        super(message);
        this.status = status;
        this.code = code;
        this.name = 'CatalogizerError';
    }
}
exports.CatalogizerError = CatalogizerError;
class AuthenticationError extends CatalogizerError {
    constructor(message = 'Authentication failed') {
        super(message, 401, 'AUTH_ERROR');
        this.name = 'AuthenticationError';
    }
}
exports.AuthenticationError = AuthenticationError;
class NetworkError extends CatalogizerError {
    constructor(message = 'Network request failed') {
        super(message, 0, 'NETWORK_ERROR');
        this.name = 'NetworkError';
    }
}
exports.NetworkError = NetworkError;
class ValidationError extends CatalogizerError {
    constructor(message = 'Validation failed') {
        super(message, 400, 'VALIDATION_ERROR');
        this.name = 'ValidationError';
    }
}
exports.ValidationError = ValidationError;
//# sourceMappingURL=index.js.map