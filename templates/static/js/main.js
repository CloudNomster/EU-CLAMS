// Main JavaScript for EU-CLAMS Web Interface

document.addEventListener('DOMContentLoaded', function() {
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
    darkModeToggle.addEventListener('click', function() {
        document.body.classList.toggle('dark-mode');
        if (document.body.classList.contains('dark-mode')) {
            localStorage.setItem('darkMode', 'true');
            darkModeToggle.textContent = '☀️';
        } else {
            localStorage.setItem('darkMode', 'false');
            darkModeToggle.textContent = '🌙';
        }
    });
    
    // Immediately refresh data when the page loads
    refreshData();
    
    // Then add periodic refresh functionality (every 30 seconds)
    setInterval(refreshData, 30000);
});

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
      // Update last updated time with browser-localized format
    document.getElementById('last-updated').textContent = new Date().toLocaleString();
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
        
        // Format the timestamp as a localized date using the browser
        const date = new Date(global.timestamp);
        timeCell.textContent = date.toLocaleString();
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
        
        // Format the timestamp as a localized date using the browser
        const date = new Date(hof.timestamp);
        timeCell.textContent = date.toLocaleString();
        typeCell.textContent = hof.type;
        targetCell.textContent = hof.target;
        valueCell.textContent = hof.value;    }
}
