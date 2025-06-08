import { useEffect, useState, useCallback } from 'react';

function useFetch<T>(url: string) {
    const [data, setData] = useState<T | null>(null);
    const [loading, setLoading] = useState<boolean>(true);
    const [error, setError] = useState<string | null>(null);

    const fetchData = useCallback(async () => {
        setLoading(true);  // Ensure loading is true when refetching
        setError(null);    // Reset error before retry
        try {
            const username = 'admin';
            const password = 'secret';
            const headers = new Headers();
            headers.set('Authorization', 'Basic ' + btoa(username + ':' + password));

            const response = await fetch(url, { headers });
            if (!response.ok) 
                throw new Error('Network error');

            const json: T = await response.json();
            setData(json);
        } catch (err: unknown) {
            if (err instanceof Error) {
                setError(err.message);
            } else {
                setError('Unknown error');
            }
        } finally {
            setLoading(false);
        }
    }, [url]);

    useEffect(() => {
        fetchData();
    }, [fetchData]);

    return { data, loading, error, refetch: fetchData };
}

export default useFetch;
