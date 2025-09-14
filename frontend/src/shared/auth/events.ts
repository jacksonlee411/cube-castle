export const AUTH_UNAUTHORIZED = 'auth:unauthorized';

class AuthEventBus extends EventTarget {
  emitUnauthorized() {
    this.dispatchEvent(new Event(AUTH_UNAUTHORIZED));
  }
}

export const authEvents = new AuthEventBus();

