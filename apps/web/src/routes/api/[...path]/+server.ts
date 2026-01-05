// Proxy all /api/* requests to the backend API
// This keeps cookies on the same domain and avoids CORS issues
import { env } from '$env/dynamic/private';
import type { RequestHandler, RequestEvent } from '@sveltejs/kit';

const BACKEND_URL = env.API_URL || 'http://localhost:8080';

type ProxyParams = { path: string };

async function proxyRequest(event: RequestEvent<ProxyParams>): Promise<Response> {
  const { request, params, cookies } = event;
  const url = new URL(request.url);
  const backendUrl = `${BACKEND_URL}/api/${params.path}${url.search}`;

  // Forward cookies to backend
  // Include both prefixed (production) and unprefixed (development) cookie names
  const headers = new Headers(request.headers);
  const cookiesToForward = ['auth_token', '__Host-auth_token', 'oauth_state', '__Host-oauth_state'];
  const cookieParts: string[] = [];

  for (const name of cookiesToForward) {
    const value = cookies.get(name);
    if (value) {
      cookieParts.push(`${name}=${value}`);
    }
  }

  if (cookieParts.length > 0) {
    headers.set('Cookie', cookieParts.join('; '));
  }

  // Remove host header (will be set by fetch)
  headers.delete('host');

  const backendResponse = await fetch(backendUrl, {
    method: request.method,
    headers,
    body: request.body,
    // @ts-expect-error - duplex is needed for streaming body
    duplex: 'half',
    redirect: 'manual' // Don't follow redirects, pass them through
  });

  // Build response headers, forwarding Set-Cookie from backend
  const responseHeaders = new Headers();

  // Forward specific headers we care about
  const headersToForward = ['content-type', 'cache-control', 'location'];
  for (const header of headersToForward) {
    const value = backendResponse.headers.get(header);
    if (value) {
      responseHeaders.set(header, value);
    }
  }

  // Forward Set-Cookie headers (can be multiple)
  const setCookies = backendResponse.headers.getSetCookie();
  for (const cookie of setCookies) {
    responseHeaders.append('Set-Cookie', cookie);
  }

  return new Response(backendResponse.body, {
    status: backendResponse.status,
    headers: responseHeaders
  });
}

export const GET: RequestHandler<ProxyParams> = async (event) => {
  return proxyRequest(event);
};

export const POST: RequestHandler<ProxyParams> = async (event) => {
  return proxyRequest(event);
};

export const PUT: RequestHandler<ProxyParams> = async (event) => {
  return proxyRequest(event);
};

export const PATCH: RequestHandler<ProxyParams> = async (event) => {
  return proxyRequest(event);
};

export const DELETE: RequestHandler<ProxyParams> = async (event) => {
  return proxyRequest(event);
};
