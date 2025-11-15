import fs from 'node:fs';
import path from 'node:path';
import type { Page, Request, Response } from '@playwright/test';

type NetworkRecord = {
  url: string;
  method: string;
  status?: number;
  type?: 'graphql' | 'rest' | 'other';
  startedAt: number;
  finishedAt?: number;
};

const outDir = path.resolve(__dirname, '../../../..', 'logs', 'plan240', 'B');

export async function installNetworkCapture(page: Page, runName: string): Promise<() => Promise<void>> {
  const records: NetworkRecord[] = [];

  const inferType = (url: string, method: string): NetworkRecord['type'] => {
    if (url.includes('/graphql') && method === 'POST') return 'graphql';
    if (url.includes('/api/')) return 'rest';
    return 'other';
  };

  const onRequest = (req: Request) => {
    records.push({
      url: req.url(),
      method: req.method(),
      type: inferType(req.url(), req.method()),
      startedAt: Date.now(),
    });
  };

  const onResponse = async (res: Response) => {
    try {
      const url = res.url();
      const method = res.request().method();
      const status = res.status();
      for (let i = records.length - 1; i >= 0; i--) {
        const r = records[i];
        if (!r.finishedAt && r.url === url && r.method === method) {
          r.status = status;
          r.finishedAt = Date.now();
          break;
        }
      }
    } catch {
      /* ignore */
    }
  };

  page.on('request', onRequest);
  page.on('response', onResponse);

  return async () => {
    try {
      fs.mkdirSync(outDir, { recursive: true });
      const file = path.join(outDir, `network-requests-${runName}-${Date.now()}.json`);
      fs.writeFileSync(file, JSON.stringify(records, null, 2), 'utf8');
    } catch {
      /* ignore */
    } finally {
      page.off('request', onRequest);
      page.off('response', onResponse);
    }
  };
}

