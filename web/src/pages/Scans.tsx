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

const Stats = () => {
    const { data, loading, error } = useFetch<Scan[]>('http://localhost:8080/api/scans');

    if (loading) return <p>Loading...</p>;
    if (error) return <p>Error: {error}</p>;
    if (!data || data.length === 0) return <div>No scans found</div>;

    return(
        <div className="flex flex-col gap-2">
            <div>
                <h3 className="font-bold text-xl mb-6">Scans History</h3>
            </div>
            {
                data.map(scan => {
                    return(
                        <div key={scan.id} className="flex flex-col md:flex-row gap-10 text-xs rounded-2xl border px-3 py-2 w-fit">
                            <div>
                                <p className="font-bold">Scan #{scan.id}</p>
                                <p>{formatDate(scan.date)}</p>
                            </div>
                            <div className="flex flex-col gap-2">
                                {
                                    scan.hosts ?
                                    scan.hosts.map(host => {
                                        return(
                                            <div key={scan.date + host.address}>
                                                <p className="font-semibold">{host.address}</p>
                                                <div className="p-2">
                                                    {
                                                        host.ports &&
                                                        host.ports.map(port => {
                                                            return(
                                                                <div key={host.address + port.port_num} className="flex gap-2 p-0.5">
                                                                    <div>{port.port_num}/{port.protocol}</div>
                                                                    <div>{port.service_name}</div>

                                                                    {
                                                                    port.vulnerabilities &&
                                                                    <div className="flex flex-col gap-1">
                                                                        <div className="text-red-500 font-semibold">
                                                                            {port.vulnerabilities.length} vulnerabilities
                                                                        </div>
                                                                        <div>
                                                                            {
                                                                                port.vulnerabilities.slice(0, 5).map(vuln => {
                                                                                    return(
                                                                                        <div key={host.address + port.port_num + vuln.cve}>
                                                                                            <p>{vuln.score} <a href={vuln.url} target="_blank">{vuln.cve}</a></p>
                                                                                        </div>
                                                                                    )
                                                                                })
                                                                            }
                                                                        </div>
                                                                    </div>
                                                                    }
                                                                    
                                                                </div>
                                                            )
                                                        })
                                                    }
                                                </div>
                                            </div>
                                        )
                                    })
                                    :
                                    <p className="text-gray-500">Scan in progress</p>
                                }
                            </div>
                        </div>
                    )
                })
            }
        </div>
    )
}

export default Stats