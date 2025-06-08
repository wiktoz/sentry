import { useState, useEffect } from "react"
import useFetch from "../hooks/useFetch"

interface ConfigInterface {
    email: string,
    scan_frequency: number,
    target: string
}

const Config = () => {
    const [email, setEmail] = useState<string>("")
    const [freq, setFreq] = useState<string>("")
    const [network, setNetwork] = useState<string>("")

    const { data, loading, error, refetch } = useFetch<ConfigInterface>('http://localhost:8080/api/config');

    useEffect(() => {
        if (!loading && data) {
            setEmail(data.email)
            setFreq(String(data.scan_frequency))
            setNetwork(data.target)
        }
    }, [loading, data]);

    if (loading) return <p>Loading...</p>;
    if (error) return <p>Error: {error}</p>;
    if (!data) return null;

    const updateConfig = async () => {
        try {
            const username = 'admin';
            const password = 'secret';
            const headers = new Headers();
            headers.set('Authorization', 'Basic ' + btoa(username + ':' + password));
            headers.set('Content-Type', 'application/json');

            const response = await fetch('http://localhost:8080/api/config', {
                method: 'PUT',
                headers,
                body: JSON.stringify({
                    email: email,
                    scan_frequency: Number(freq),
                    target: network,
                }),
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            // Re-fetch updated config
            if (typeof refetch === 'function') {
                refetch();
            }

        } catch (err) {
            console.error('Update failed:', err);
        }
    }

    return(
        <div>
            <div>
                <h3 className="font-bold text-xl mb-6">System Config</h3>
            </div>
            <div className="flex flex-col gap-2">
                <div className="flex flex-col w-full md:w-72">
                    <label htmlFor="email" className="text-xs p-1">e-mail for raporting</label>
                    <input id="email" type="text" placeholder="e-mail" className="border border-black px-3 py-1 rounded-xl focus:outline-0" onChange={e => setEmail(e.target.value)} value={email}/>
                </div>
                <div className="flex flex-col w-full md:w-72">
                    <label htmlFor="scan_freq" className="text-xs p-1">auto-scan frequency in seconds</label>
                    <input id="scan_freq" type="number" placeholder="scan frequency (s)" className="border border-black px-3 py-1 rounded-xl focus:outline-0" onChange={e => setFreq(e.target.value)} value={freq}/>
                </div>
                <div className="flex flex-col w-full md:w-72">
                    <label htmlFor="network" className="text-xs p-1">scanned network</label>
                    <input id="network" type="text" placeholder="scanned network" className="border border-black px-3 py-1 rounded-xl focus:outline-0" onChange={e => setNetwork(e.target.value)} value={network}/>
                </div>
                <div className="bg-black text-white font-semibold rounded-xl px-3 py-2 w-full md:w-72 text-sm text-center cursor-pointer my-2" onClick={() => updateConfig()}>
                    Update Config
                </div>
            </div>
            
        </div>
    )
}

export default Config