// Initialize Bootstrap dropdown
var dropdownElementList = [].slice.call(document.querySelectorAll('.dropdown-toggle'));
var dropdownList = dropdownElementList.map(function (dropdownToggleEl) {
  return new bootstrap.Dropdown(dropdownToggleEl);
});

// Elements selection
const form = document.getElementById('download-form');
const resultDiv = document.getElementById('result');
const spinner = document.getElementsByClassName('spinner')[0];
const videoDetailsDiv = document.getElementById('video-details');
const videoDetailsTable = document.getElementById('video-details-table');
const downloadButton = document.getElementById('download-btn');

// Handle form submission
form.addEventListener('submit', async (event) => {
  event.preventDefault();

  const urlInput = document.getElementById('url-input');
  const url = urlInput.value.trim();
  clearUI(); // Clear the UI for the new request

  if (!url) {
    resultDiv.innerHTML = '<p>Please enter a valid YouTube URL.</p>';
    return;
  }

  try {
    await fetchVideoDetails(url);
    const downloadToken = await initiateDownload(url);
    await pollDownload(downloadToken);
  } catch (error) {
    console.error('Error:', error);
    displayError(error.message || 'An unexpected error occurred');
  }
});

// Fetch video details and update UI
async function fetchVideoDetails(url) {
  spinner.style.display = 'flex';
  const response = await fetch('/info?url=' + encodeURIComponent(url));
  if (!response.ok) throw new Error('Failed to fetch video details');

  const videoDetails = await response.json();
  displayVideoDetails(videoDetails.title, videoDetails.duration);
}

// Initiate download and return token
async function initiateDownload(url) {
  const response = await fetch('/download?url=' + encodeURIComponent(url));
  if (!response.ok) {
    const errorMessage = await response.text();
    throw new Error(errorMessage);
  }

  const data = await response.json();
  return data.token;
}

// Poll the server for download progress and handle download
async function pollDownload(downloadToken) {
  resultDiv.innerHTML = '<p>Downloading...</p>';
  const checkDownload = async () => {
    const response = await fetch('/progress?token=' + encodeURIComponent(downloadToken));

    if (response.status === 202) {
      setTimeout(checkDownload, 5000); // Poll every 5 seconds
    } else if (response.status === 200) {
      await downloadFile(downloadToken);
    } else {
      throw new Error('Download failed');
    }
  };

  checkDownload();
}

// Perform the actual file download
async function downloadFile(downloadToken) {
  const response = await fetch('/downloadFile?token=' + encodeURIComponent(downloadToken));
  spinner.style.display = 'none';

  if (!response.ok) {
    const errorMessage = await response.text();
    throw new Error(errorMessage);
  }

  const blob = await response.blob();
  createDownloadLink(blob);
}

// Create and handle download link
function createDownloadLink(blob) {
  const downloadLink = document.createElement('a');
  downloadLink.href = URL.createObjectURL(blob);
  downloadLink.download = 'downloadedFile'; // Set your desired file name

  const downloadButton = document.getElementById('download-btn');
  downloadButton.style.display = 'inline-block';
  downloadButton.onclick = () => {
    document.body.appendChild(downloadLink);
    downloadLink.click();
    document.body.removeChild(downloadLink);
    resultDiv.innerHTML = '<p>Download completed.</p>';
  };

  resultDiv.innerHTML = '<p>Download is ready.</p>';
  spinner.style.display = 'none';
}

// Display video details on the UI
function displayVideoDetails(title, duration) {
  videoDetailsDiv.style.display = 'flex';
  const newRow = document.createElement('tr');
  newRow.innerHTML = `<td>${title}</td><td>${duration}</td>`;
  videoDetailsTable.appendChild(newRow);
}

// Display error messages
function displayError(message) {
  resultDiv.innerHTML = `<p>Error: ${message}</p>`;
  spinner.style.display = 'none';
}

// Clear UI elements
function clearUI() {
  spinner.style.display = 'none';
  downloadButton.style.display = 'none';
  resultDiv.innerHTML = "";
  videoDetailsTable.innerHTML = "";
  videoDetailsDiv.style.display = 'none';
}

// Read URL parameters and auto-submit form if applicable
const queryString = window.location.search;
const urlParams = new URLSearchParams(queryString);
if (urlParams.has('v')) {
  document.getElementById('url-input').value = `https://www.youtube.com/watch?v=${urlParams.get('v')}`;
  form.dispatchEvent(new Event('submit'));
}
