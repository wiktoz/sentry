import useFetch from "../hooks/useFetch";

interface Scan {
    date: string,
    hosts: [{
        address: string,
        ports: {
            port_num: number,
            service_name: string,
            protocol: string,
            state: string
        }
    }]
}

const Stats = () => {
    const { data, loading, error } = useFetch<Scan[]>('http://localhost:8080/api/scans');

    if (loading) return <p>Loading...</p>;
    if (error) return <p>Error: {error}</p>;
    if (!data) return null;

    return(
        <div>
            {
                data.map(scan => {
                    return(
                        <div key={scan.date} className="flex gap-2 text-xs">
                            <div>
                                {scan.date}
                            </div>
                            <div>
                                {
                                    scan.hosts.map(host => {
                                        return(
                                            <div key={scan.date + host.address}>
                                                {host.address}
                                            </div>
                                        )
                                    })
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