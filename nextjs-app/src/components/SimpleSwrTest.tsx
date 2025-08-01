import React from 'react';
import useSWR from 'swr';

const simpleFetcher = async (url: string) => {
  console.log('🔬 Simple Test Fetcher Called:', url);
  const response = await fetch(url);
  const data = await response.json();
  console.log('🔬 Simple Test Fetcher Success:', data);
  return data;
};

export function SimpleSwrTest() {
  console.log('🔬 SimpleSwrTest Component Rendered');
  
  const { data, error, isLoading } = useSWR(
    '/api/employees?page=1&page_size=3',
    simpleFetcher,
    {
      revalidateOnFocus: false,
      revalidateOnReconnect: false,
      revalidateIfStale: false,
    }
  );
  
  console.log('🔬 Simple SWR State:', { data: !!data, error: !!error, isLoading });
  
  return (
    <div>
      <h3>Simple SWR Test</h3>
      <div>Loading: {isLoading ? 'Yes' : 'No'}</div>
      <div>Error: {error ? error.message : 'None'}</div>
      <div>Data: {data ? `${data.employees?.length || 0} employees` : 'None'}</div>
    </div>
  );
}