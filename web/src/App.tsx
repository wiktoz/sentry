import './App.css'
import { useState, useEffect } from 'react'
import Home from './pages/Home'
import Config from './pages/Config'
import Scans from './pages/Scans'

export const Page = {
	Home: "Home",
	Scans: "Scans",
	Config: "Config"
} as const;

type Page = typeof Page[keyof typeof Page];

const pageComponents: Record<Page, React.FC> = {
	[Page.Home]: Home,
  	[Page.Scans]: Scans,
  	[Page.Config]: Config,
};

const STORAGE_KEY = 'lastPageBookmark';

function App() {
	// Initialize page from localStorage if available
	const [pageOpen, setPageOpen] = useState<Page>(() => {
		const savedPage = localStorage.getItem(STORAGE_KEY);
		if (savedPage && Object.values(Page).includes(savedPage as Page)) {
			return savedPage as Page;
		}
		return Page.Home;
	});

	useEffect(() => {
		// Save page to localStorage when it changes
		localStorage.setItem(STORAGE_KEY, pageOpen);
	}, [pageOpen]);

	const CurrentPageComponent = pageComponents[pageOpen];

	return (
    	<div className='m-8 flex flex-col gap-8'>
			<nav className='flex gap-2'>
				{Object.values(Page).map((pageValue) => (
					<div
						key={pageValue}
						onClick={() => setPageOpen(pageValue)}
						className={"cursor-pointer transition-all font-semibold rounded-3xl px-4 py-1.5" 
							+ " " + (pageOpen === pageValue && "bg-black text-white")}
					>
						{pageValue}
					</div>
				))}
			</nav>

			<main className='w-full rounded-3xl p-2'>
				<CurrentPageComponent/>
			</main>
    	</div>
	)
}

export default App
