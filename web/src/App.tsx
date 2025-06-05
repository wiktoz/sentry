import './App.css'
import { useState } from 'react'
import Home from './pages/Home'
import Config from './pages/Config'
import Stats from './pages/Stats'

export const Page = {
	Home: "Home",
	Stats: "Stats",
	Config: "Config"
} as const;

type Page = typeof Page[keyof typeof Page];

const pageComponents: Record<Page, React.FC> = {
	[Page.Home]: Home,
  	[Page.Stats]: Stats,
  	[Page.Config]: Config,
};

function App() {
	const [pageOpen, setPageOpen] = useState<Page>(Page.Home)

	const CurrentPageComponent = pageComponents[pageOpen];

	return (
    	<div className='m-8 flex flex-col gap-8'>
			<nav className='flex gap-2'>
				{Object.values(Page).map((pageValue) => (
					<div
						key={pageValue}
						onClick={() => setPageOpen(pageValue)}
						className={"border border-black cursor-pointer transition-all font-semibold rounded-3xl px-4 py-1.5" 
							+ " " + (pageOpen === pageValue && "bg-black text-white")}
					>
						{pageValue}
					</div>
				))}
			</nav>

			<main className='w-full border border-black rounded-3xl p-8'>
				<CurrentPageComponent/>
			</main>
    	</div>
	)
}

export default App
