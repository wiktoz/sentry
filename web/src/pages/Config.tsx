const Config = () => {
    return(
        <div>
            <div>
                <h3 className="font-bold text-xl mb-6">System Config</h3>
            </div>
            <div className="flex flex-col gap-2">
                <div className="flex flex-col w-full md:w-72">
                    <label htmlFor="email" className="text-xs p-1">e-mail for raporting</label>
                    <input id="email" type="text" placeholder="e-mail" className="border border-black px-3 py-1 rounded-xl focus:outline-0"/>
                </div>
                <div className="flex flex-col w-full md:w-72">
                    <label htmlFor="email" className="text-xs p-1">auto-scan frequency in seconds</label>
                    <input id="scan_freq" type="number" placeholder="scan frequency (s)" className="border border-black px-3 py-1 rounded-xl focus:outline-0"/>
                </div>
                <div className="flex flex-col w-full md:w-72">
                    <label htmlFor="email" className="text-xs p-1">scanned network</label>
                    <input id="network" type="text" placeholder="scanned network" className="border border-black px-3 py-1 rounded-xl focus:outline-0"/>
                </div>
                <div className="bg-black text-white font-semibold rounded-xl px-3 py-2 w-full md:w-72 text-sm text-center cursor-pointer my-2">
                    Update Config
                </div>
            </div>
            
        </div>
    )
}

export default Config