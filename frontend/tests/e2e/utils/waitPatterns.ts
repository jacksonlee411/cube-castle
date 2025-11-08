import type { Page, Response } from '@playwright/test';

type UrlMatcher = string | RegExp;

interface WaitOptions {
  timeout?: number;
}

interface GraphQLWaitOptions extends WaitOptions {
  urlMatcher?: UrlMatcher;
}

const DEFAULT_TIMEOUT = 15_000;
const DEFAULT_GRAPHQL_MATCHER: UrlMatcher = /\/graphql/;

const matchesUrl = (url: string, matcher: UrlMatcher): boolean =>
  typeof matcher === 'string' ? url.includes(matcher) : matcher.test(url);

const getRequestBody = (response: Response): string | undefined => {
  try {
    return response.request().postData() ?? undefined;
  } catch (_error) {
    return undefined;
  }
};

const bodyIncludesOperation = (body: string | undefined, operation: string | RegExp): boolean => {
  if (!body) {
    return false;
  }
  return typeof operation === 'string' ? body.includes(operation) : operation.test(body);
};

export const waitForPageReady = async (page: Page, options?: WaitOptions): Promise<void> => {
  const timeout = options?.timeout ?? DEFAULT_TIMEOUT;
  await page.waitForLoadState('domcontentloaded', { timeout });
  try {
    await page.waitForLoadState('networkidle', { timeout });
  } catch (_error) {
    // 某些页面不会进入 networkidle 状态，忽略该异常即可
  }
};

export const waitForNavigation = async (
  page: Page,
  expectedUrl: UrlMatcher,
  options?: WaitOptions,
): Promise<void> => {
  const timeout = options?.timeout ?? DEFAULT_TIMEOUT;
  await page.waitForURL(expectedUrl, { timeout });
  await waitForPageReady(page, { timeout });
};

export const waitForGraphQL = async (
  page: Page,
  operationName: string | RegExp,
  options?: GraphQLWaitOptions,
): Promise<void> => {
  const timeout = options?.timeout ?? DEFAULT_TIMEOUT;
  const urlMatcher = options?.urlMatcher ?? DEFAULT_GRAPHQL_MATCHER;

  await page.waitForResponse(
    response =>
      response.request().method() === 'POST' &&
      matchesUrl(response.url(), urlMatcher) &&
      bodyIncludesOperation(getRequestBody(response), operationName),
    { timeout },
  );
};
