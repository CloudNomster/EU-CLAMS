// Main JavaScript for EU-CLAMS Web Interface

document.addEventListener('DOMContentLoaded', function () {
    // Add dark mode toggle
    const darkModeToggle = document.createElement('div');
    darkModeToggle.className = 'dark-mode-toggle';
    darkModeToggle.textContent = '🌙';
    darkModeToggle.title = 'Toggle Dark Mode';
    document.body.appendChild(darkModeToggle);

    // Check for saved dark mode preference
    if (localStorage.getItem('darkMode') === 'true') {
        document.body.classList.add('dark-mode');
        darkModeToggle.textContent = '☀️';
    }

    // Dark mode toggle function
    darkModeToggle.addEventListener('click', function () {
        document.body.classList.toggle('dark-mode');
        if (document.body.classList.contains('dark-mode')) {
            localStorage.setItem('darkMode', 'true');
            darkModeToggle.textContent = '☀️';
        } else {
            localStorage.setItem('darkMode', 'false');
            darkModeToggle.textContent = '🌙';
        }
    });

    // Initialize table sorting
    initTableSorting();

    // Load initial data when the page loads
    refreshData();

    // Periodic refresh removed to preserve sort state
    // WebSocket will handle live updates for new events
});


// Helper function to format date in 24-hour format
function formatDateTime(date) {
    return date.toLocaleString('en-GB', {
        hour12: false,
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit'
    });
}

// Function to refresh data
function refreshData() {
    // Fetch updated stats
    fetch('/api/stats')
        .then(response => response.json())
        .then(stats => {
            updateStats(stats);
        })
        .catch(error => console.error('Error fetching stats:', error));

    // Fetch updated globals
    fetch('/api/globals')
        .then(response => response.json())
        .then(globals => {
            updateGlobals(globals);
        })
        .catch(error => console.error('Error fetching globals:', error));

    // Fetch updated HOFs
    fetch('/api/hofs')
        .then(response => response.json())
        .then(hofs => {
            updateHofs(hofs);
        })
        .catch(error => console.error('Error fetching HOFs:', error));
    // Update last updated time in 24-hour format
    document.getElementById('last-updated').textContent = formatDateTime(new Date());
}

// Function to update stats display
function updateStats(stats) {
    document.getElementById('total-globals').textContent = stats.TotalGlobals;
    document.getElementById('total-hofs').textContent = stats.TotalHofs;
    document.getElementById('total-value').textContent = stats.TotalValue.toFixed(2);
    document.getElementById('highest-value').textContent = stats.HighestValue.toFixed(2);

    // Update globals by type
    const typeTable = document.getElementById('globals-by-type');
    typeTable.innerHTML = '';
    for (const [type, count] of Object.entries(stats.ByType)) {
        const row = typeTable.insertRow();
        const typeCell = row.insertCell(0);
        const countCell = row.insertCell(1);
        typeCell.textContent = type;
        countCell.textContent = count;
    }

    // Update globals by location
    const locationTable = document.getElementById('globals-by-location');
    locationTable.innerHTML = '';
    for (const [location, count] of Object.entries(stats.ByLocation)) {
        const row = locationTable.insertRow();
        const locationCell = row.insertCell(0);
        const countCell = row.insertCell(1);
        locationCell.textContent = location;
        countCell.textContent = count;
    }
}

// Function to update globals table
function updateGlobals(globals) {
    const table = document.getElementById('latest-globals');
    table.innerHTML = '';

    // Check if globals is an array or convert it to one if possible
    const globalsArray = Array.isArray(globals) ? globals :
        (globals && typeof globals === 'object') ? Object.values(globals) : [];

    // If globalsArray is empty, maybe add a "No data" row
    if (globalsArray.length === 0) {
        const row = table.insertRow();
        const cell = row.insertCell(0);
        cell.colSpan = 4;
        cell.textContent = "No global data available";
        cell.className = "no-data";
        return;
    }
    for (const global of globalsArray) {
        const row = table.insertRow();
        const timeCell = row.insertCell(0);
        const typeCell = row.insertCell(1);
        const targetCell = row.insertCell(2);
        const valueCell = row.insertCell(3);

        // Format the timestamp in 24-hour format
        const date = new Date(global.timestamp);
        timeCell.textContent = formatDateTime(date);
        typeCell.textContent = global.type;
        targetCell.textContent = global.target;
        valueCell.textContent = global.value;
    }
}

// Function to update HOFs table
function updateHofs(hofs) {
    const table = document.getElementById('latest-hofs');
    table.innerHTML = '';

    // Check if hofs is an array or convert it to one if possible
    const hofsArray = Array.isArray(hofs) ? hofs :
        (hofs && typeof hofs === 'object') ? Object.values(hofs) : [];

    // If hofsArray is empty, maybe add a "No data" row
    if (hofsArray.length === 0) {
        const row = table.insertRow();
        const cell = row.insertCell(0);
        cell.colSpan = 4;
        cell.textContent = "No HOF data available";
        cell.className = "no-data";
        return;
    }
    for (const hof of hofsArray) {
        const row = table.insertRow();
        const timeCell = row.insertCell(0);
        const typeCell = row.insertCell(1);
        const targetCell = row.insertCell(2);
        const valueCell = row.insertCell(3);

        // Format the timestamp in 24-hour format
        const date = new Date(hof.timestamp);
        timeCell.textContent = formatDateTime(date);
        typeCell.textContent = hof.type;
        targetCell.textContent = hof.target;
        valueCell.textContent = hof.value;
    }
}

// Sort state tracking
const sortState = {
    globals: { column: null, ascending: true },
    hofs: { column: null, ascending: true }
};

// Initialize table sorting
function initTableSorting() {
    // Add click event listeners to all sortable headers
    document.querySelectorAll('th.sortable').forEach(header => {
        header.addEventListener('click', function () {
            const table = this.getAttribute('data-table');
            const sortBy = this.getAttribute('data-sort');
            sortTable(table, sortBy);
        });
    });
}

// Sort table by specified column
function sortTable(tableName, sortBy) {
    const state = sortState[tableName];

    // Toggle sort direction if clicking the same column
    if (state.column === sortBy) {
        state.ascending = !state.ascending;
    } else {
        state.column = sortBy;
        state.ascending = true;
    }

    // Get the table body
    const tbody = document.getElementById(tableName === 'globals' ? 'latest-globals' : 'latest-hofs');
    const rows = Array.from(tbody.getElementsByTagName('tr'));

    // Sort rows based on column and direction
    rows.sort((a, b) => {
        let aVal, bVal;

        if (sortBy === 'time') {
            // Get timestamp from first cell
            const aTime = new Date(a.cells[0].textContent);
            const bTime = new Date(b.cells[0].textContent);
            aVal = aTime.getTime();
            bVal = bTime.getTime();
        } else if (sortBy === 'value') {
            // Get value from last cell (remove any non-numeric characters)
            aVal = parseFloat(a.cells[3].textContent.replace(/[^0-9.-]/g, '')) || 0;
            bVal = parseFloat(b.cells[3].textContent.replace(/[^0-9.-]/g, '')) || 0;
        }

        if (state.ascending) {
            return aVal - bVal;
        } else {
            return bVal - aVal;
        }
    });

    // Clear and rebuild table body
    tbody.innerHTML = '';
    rows.forEach(row => tbody.appendChild(row));

    // Update sort indicators
    updateSortIndicators(tableName, sortBy, state.ascending);
}


// Update visual indicators for sorted column
function updateSortIndicators(tableName, sortBy, ascending) {
    // Clear all indicators for this table
    document.querySelectorAll(`th[data-table="${tableName}"] .sort-indicator`).forEach(indicator => {
        indicator.textContent = '';
    });

    // Set indicator for active column
    const activeHeader = document.querySelector(`th[data-table="${tableName}"][data-sort="${sortBy}"] .sort-indicator`);
    if (activeHeader) {
        activeHeader.textContent = ascending ? ' ▲' : ' ▼';
    }
}
