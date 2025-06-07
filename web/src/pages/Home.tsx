const Home = () => {
    return(
        <div className="flex flex-col gap-8">
            <div>
                <div>
                    <h3 className="font-bold text-xl mb-4">Home</h3>
                </div>
                <div className="bg-black text-white font-semibold rounded-xl px-3 py-2 w-full md:w-72 text-sm text-center cursor-pointer my-2">
                    Run Scan
                </div>
            </div>
            <div>
                <div>
                    <h3 className="font-bold text-xl mb-4">Notifications</h3>
                </div>
                <div>
                    <h2><span className="font-semibold text-red-500">Warning!</span> You have <span className="text-red-500">378 vulnerabilities</span> in your network.</h2>
                    <p className="text-gray-800 text-xs my-2">Go to <span className="font-semibold">Scans</span> page to see more details.</p>
                </div>
            </div>
        </div>
    )
}

export default Home