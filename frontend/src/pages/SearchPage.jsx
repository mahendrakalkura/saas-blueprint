import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Search, Filter, TrendingUp } from 'lucide-react'
import SearchComponent from '../components/Search'

export default function SearchPage() {
  const navigate = useNavigate()
  const [selectedFilters, setSelectedFilters] = useState({
    entityTypes: [],
  })

  const handleResultClick = (result) => {
    // Navigate based on entity type
    switch (result.entity_type) {
      case 'user':
        navigate(`/users/${result.entity_id}`)
        break
      case 'organization':
        navigate(`/organizations/${result.entity_id}`)
        break
      case 'notification':
        navigate(`/notifications`)
        break
      default:
        console.log('Unknown entity type:', result.entity_type)
    }
  }

  const toggleFilter = (filterType, value) => {
    setSelectedFilters((prev) => {
      const currentFilters = prev[filterType] || []
      const newFilters = currentFilters.includes(value)
        ? currentFilters.filter((f) => f !== value)
        : [...currentFilters, value]

      return {
        ...prev,
        [filterType]: newFilters,
      }
    })
  }

  const popularSearches = [
    'notifications',
    'settings',
    'billing',
    'members',
    'files',
  ]

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900 py-8">
      <div className="max-w-4xl mx-auto px-4">
        {/* Header */}
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-gradient-to-br from-blue-500 to-purple-600 mb-4">
            <Search className="w-8 h-8 text-white" />
          </div>
          <h1 className="text-3xl font-bold text-gray-900 dark:text-white mb-2">
            Search
          </h1>
          <p className="text-gray-600 dark:text-gray-400">
            Search across users, organizations, and notifications
          </p>
        </div>

        {/* Search Bar */}
        <div className="mb-8 flex justify-center">
          <SearchComponent
            onResultClick={handleResultClick}
            placeholder="Search for anything..."
            autoFocus
          />
        </div>

        {/* Filters */}
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-4 mb-8">
          <div className="flex items-center gap-2 mb-3">
            <Filter className="w-4 h-4 text-gray-500" />
            <h2 className="text-sm font-semibold text-gray-900 dark:text-white">
              Filter by type
            </h2>
          </div>

          <div className="flex flex-wrap gap-2">
            {[
              { value: 'user', label: 'Users', color: 'blue' },
              { value: 'organization', label: 'Organizations', color: 'purple' },
              { value: 'notification', label: 'Notifications', color: 'green' },
            ].map((filter) => (
              <button
                key={filter.value}
                onClick={() => toggleFilter('entityTypes', filter.value)}
                className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                  selectedFilters.entityTypes.includes(filter.value)
                    ? `bg-${filter.color}-100 text-${filter.color}-700 dark:bg-${filter.color}-900 dark:text-${filter.color}-300 border-2 border-${filter.color}-500`
                    : 'bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-300 border-2 border-transparent hover:border-gray-300 dark:hover:border-gray-600'
                }`}
              >
                {filter.label}
              </button>
            ))}
          </div>
        </div>

        {/* Popular Searches */}
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-4">
          <div className="flex items-center gap-2 mb-3">
            <TrendingUp className="w-4 h-4 text-gray-500" />
            <h2 className="text-sm font-semibold text-gray-900 dark:text-white">
              Popular searches
            </h2>
          </div>

          <div className="flex flex-wrap gap-2">
            {popularSearches.map((search) => (
              <button
                key={search}
                onClick={() => {
                  // Trigger search with this term
                  // You would need to pass this to the SearchComponent
                  console.log('Popular search clicked:', search)
                }}
                className="px-3 py-1.5 rounded-full text-sm text-gray-600 dark:text-gray-400 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors"
              >
                {search}
              </button>
            ))}
          </div>
        </div>

        {/* Search Tips */}
        <div className="mt-8 bg-blue-50 dark:bg-blue-900/20 rounded-lg p-4 border border-blue-200 dark:border-blue-800">
          <h3 className="text-sm font-semibold text-blue-900 dark:text-blue-300 mb-2">
            Search Tips
          </h3>
          <ul className="text-xs text-blue-800 dark:text-blue-400 space-y-1">
            <li>• Use quotes for exact phrases: "project manager"</li>
            <li>• Use OR to search for multiple terms: design OR development</li>
            <li>• Use - to exclude terms: notification -email</li>
            <li>• Use keyboard shortcuts: ↑↓ to navigate, Enter to select, Esc to close</li>
          </ul>
        </div>
      </div>
    </div>
  )
}
