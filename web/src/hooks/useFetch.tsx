import { useEffect, useState } from 'react';

function useFetch<T>(url: string) {
    const [data, setData] = useState<T | null>(null);
    const [loading, setLoading] = useState<boolean>(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        async function fetchData() {
        try {
            const response = await fetch(url);
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
        }
        fetchData();
    }, [url]);

    return { data, loading, error };
}


export default useFetch