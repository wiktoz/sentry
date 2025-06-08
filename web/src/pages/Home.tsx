import { useState } from "react";
import useFetch from "../hooks/useFetch";

interface Scan {
    id: number,
    date: string,
    hosts?: [{
        address: string,
        ports?: [{
            port_num: number,
            service_name: string,
            protocol: string,
            state: string
            vulnerabilities: [{
                cve: string,
	            description: string,
                score: string,
                url: string
            }]
        }]
    }]
}

function formatDate(rawDate:string) {
    const isoString = rawDate.replace(' ', 'T');
    const date = new Date(isoString);

    return date.toLocaleString('pl-PL', {
        year: 'numeric', month: '2-digit', day: '2-digit',
        hour: '2-digit', minute: '2-digit', second: '2-digit',
        hour12: false,
    });
}

const Home = () => {
    const [run, setRun] = useState<boolean>(false)

    const { data, loading, error } = useFetch<Scan[]>('http://localhost:8080/api/scans');

    if (loading) return <div>Loading...</div>;
    if (error) return <div>Error: {error}</div>;

    // Find newest scan by date
    const newestScan = data && data.length > 0
    ? data.reduce((latest, current) => new Date(current.date) > new Date(latest.date) ? current : latest)
    : undefined;

    const filteredScans = data?.filter(scan => 
        scan.hosts?.some(host => 
            host.ports?.some(port => 
                port.vulnerabilities && port.vulnerabilities.length > 0
            )
        )
    ) ?? [];

    const newestScanWithVulns = filteredScans.length > 0
    ? filteredScans.reduce((latest, current) => 
        new Date(current.date) > new Date(latest.date) ? current : latest
      )
    : undefined;

    // Count vulnerabilities in newest scan
    const totalVulnerabilities = newestScanWithVulns?.hosts?.reduce((acc, host) => acc + (host.ports?.reduce((pAcc, port) => pAcc + (port.vulnerabilities?.length ?? 0), 0) ?? 0), 0) ?? 0;


    const runScan = async () => {
        try {
            const username = 'admin';
            const password = 'secret';
            const headers = new Headers();
            headers.set('Authorization', 'Basic ' + btoa(username + ':' + password));
            headers.set('Content-Type', 'application/json');

            const response = await fetch('http://localhost:8080/api/scan/run', {
                method: 'GET',
                headers,
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            setRun(true)

        } catch (err) {
            console.error('Update failed:', err);
        }
    }

    return(
        <div className="flex flex-col gap-8">
            <div>
                <div>
                    <h3 className="font-bold text-xl mb-4">Home</h3>
                </div>
                <div className="bg-black text-white font-semibold rounded-xl px-3 py-2 w-full md:w-72 text-sm text-center cursor-pointer my-2" onClick={() => runScan()}>
                    Run Scan
                </div>
                {
                    run &&
                    <div className="">
                        Scan started!
                    </div>
                }
            </div>
            <div>
                <div>
                    <h3 className="font-bold text-xl mb-4">Notifications</h3>
                </div>
                <div className="mb-4">
                    <h2>newest scan: <span className="font-semibold">{newestScan ? formatDate(newestScan.date) : "No scans"}</span></h2>
                </div>
                {
                    totalVulnerabilities > 0 &&
                
                    <div>
                        <h2><span className="font-semibold text-red-500">Warning!</span> You have <span className="text-red-500">{totalVulnerabilities} vulnerabilities</span> in your network.</h2>
                        <p className="text-gray-800 text-xs my-2">Go to <span className="font-semibold">Scans</span> page to see more details.</p>
                    </div>
                }
            </div>
        </div>
    )
}

export default Home