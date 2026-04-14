import { describe, it, expect } from 'vitest';
import {
  CatalogizerError,
  AuthenticationError,
  NetworkError,
  ValidationError,
} from '../index';

describe('Error classes', () => {
  describe('CatalogizerError', () => {
    it('creates error with message only', () => {
      const error = new CatalogizerError('something failed');

      expect(error).toBeInstanceOf(Error);
      expect(error).toBeInstanceOf(CatalogizerError);
      expect(error.message).toBe('something failed');
      expect(error.name).toBe('CatalogizerError');
      expect(error.status).toBeUndefined();
      expect(error.code).toBeUndefined();
    });

    it('creates error with message and status', () => {
      const error = new CatalogizerError('not found', 404);

      expect(error.message).toBe('not found');
      expect(error.status).toBe(404);
      expect(error.code).toBeUndefined();
    });

    it('creates error with message, status, and code', () => {
      const error = new CatalogizerError('forbidden', 403, 'FORBIDDEN');

      expect(error.message).toBe('forbidden');
      expect(error.status).toBe(403);
      expect(error.code).toBe('FORBIDDEN');
    });

    it('has a proper stack trace', () => {
      const error = new CatalogizerError('test error');

      expect(error.stack).toBeDefined();
      expect(error.stack).toContain('CatalogizerError');
    });

    it('can be caught as a generic Error', () => {
      let caught: Error | null = null;

      try {
        throw new CatalogizerError('test', 500, 'SERVER_ERROR');
      } catch (e) {
        caught = e as Error;
      }

      expect(caught).toBeInstanceOf(Error);
      expect(caught).toBeInstanceOf(CatalogizerError);
      expect((caught as CatalogizerError).status).toBe(500);
      expect((caught as CatalogizerError).code).toBe('SERVER_ERROR');
    });
  });

  describe('AuthenticationError', () => {
    it('creates error with default message', () => {
      const error = new AuthenticationError();

      expect(error).toBeInstanceOf(Error);
      expect(error).toBeInstanceOf(CatalogizerError);
      expect(error).toBeInstanceOf(AuthenticationError);
      expect(error.message).toBe('Authentication failed');
      expect(error.name).toBe('AuthenticationError');
      expect(error.status).toBe(401);
      expect(error.code).toBe('AUTH_ERROR');
    });

    it('creates error with custom message', () => {
      const error = new AuthenticationError('Token expired');

      expect(error.message).toBe('Token expired');
      expect(error.status).toBe(401);
      expect(error.code).toBe('AUTH_ERROR');
    });

    it('is distinguishable from parent via instanceof', () => {
      const authError = new AuthenticationError();
      const baseError = new CatalogizerError('generic', 401);

      expect(authError).toBeInstanceOf(AuthenticationError);
      expect(baseError).not.toBeInstanceOf(AuthenticationError);
    });
  });

  describe('NetworkError', () => {
    it('creates error with default message', () => {
      const error = new NetworkError();

      expect(error).toBeInstanceOf(Error);
      expect(error).toBeInstanceOf(CatalogizerError);
      expect(error).toBeInstanceOf(NetworkError);
      expect(error.message).toBe('Network request failed');
      expect(error.name).toBe('NetworkError');
      expect(error.status).toBe(0);
      expect(error.code).toBe('NETWORK_ERROR');
    });

    it('creates error with custom message', () => {
      const error = new NetworkError('DNS resolution failed');

      expect(error.message).toBe('DNS resolution failed');
      expect(error.status).toBe(0);
      expect(error.code).toBe('NETWORK_ERROR');
    });
  });

  describe('ValidationError', () => {
    it('creates error with default message', () => {
      const error = new ValidationError();

      expect(error).toBeInstanceOf(Error);
      expect(error).toBeInstanceOf(CatalogizerError);
      expect(error).toBeInstanceOf(ValidationError);
      expect(error.message).toBe('Validation failed');
      expect(error.name).toBe('ValidationError');
      expect(error.status).toBe(400);
      expect(error.code).toBe('VALIDATION_ERROR');
    });

    it('creates error with custom message', () => {
      const error = new ValidationError('Email is required');

      expect(error.message).toBe('Email is required');
      expect(error.status).toBe(400);
      expect(error.code).toBe('VALIDATION_ERROR');
    });
  });

  describe('Error hierarchy', () => {
    it('all custom errors extend CatalogizerError', () => {
      expect(new AuthenticationError()).toBeInstanceOf(CatalogizerError);
      expect(new NetworkError()).toBeInstanceOf(CatalogizerError);
      expect(new ValidationError()).toBeInstanceOf(CatalogizerError);
    });

    it('all custom errors extend Error', () => {
      expect(new AuthenticationError()).toBeInstanceOf(Error);
      expect(new NetworkError()).toBeInstanceOf(Error);
      expect(new ValidationError()).toBeInstanceOf(Error);
    });

    it('error types are distinguishable from each other', () => {
      const authErr = new AuthenticationError();
      const netErr = new NetworkError();
      const valErr = new ValidationError();

      expect(authErr).not.toBeInstanceOf(NetworkError);
      expect(authErr).not.toBeInstanceOf(ValidationError);
      expect(netErr).not.toBeInstanceOf(AuthenticationError);
      expect(netErr).not.toBeInstanceOf(ValidationError);
      expect(valErr).not.toBeInstanceOf(AuthenticationError);
      expect(valErr).not.toBeInstanceOf(NetworkError);
    });

    it('errors serialize name and message in string conversion', () => {
      const error = new CatalogizerError('test message', 500, 'TEST');
      const str = error.toString();

      expect(str).toContain('CatalogizerError');
      expect(str).toContain('test message');
    });
  });
});
