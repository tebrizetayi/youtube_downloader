// Initialize Bootstrap dropdown
var dropdownElementList = [].slice.call(document.querySelectorAll('.dropdown-toggle'))
var dropdownList = dropdownElementList.map(function (dropdownToggleEl) {
  return new bootstrap.Dropdown(dropdownToggleEl)
});

const form = document.getElementById('download-form');
const resultDiv = document.getElementById('result');
const spinner = document.getElementsByClassName('spinner')[0];
const videoDetailsDiv = document.getElementById('video-details');
const videoDetailsDuration = document.getElementById('video-duration');
const videoDetailsTitle = document.getElementById('video-title');
const videoDetailstable = document.getElementById('video-details-table');
const downloadButton = document.getElementById('download-btn');

form.addEventListener('submit', async (event) => {
  console.log("begin")
  event.preventDefault();

  const urlInput = document.getElementById('url-input');
  const url = urlInput.value.trim();
  spinner.style.display = 'none';
  downloadButton.style.display = 'none';

  resultDiv.innerHTML = "";
  videoDetailstable.innerHTML = "";
  videoDetailsDiv.style.display = 'none';

  if (!url) {
    resultDiv.innerHTML = '<p>Please enter a valid YouTube URL.</p>';
    return;
  }

  // Show the loading spinner
  spinner.style.display = 'flex';

  const videoDetailsResponse = await fetch('/info?url=' + encodeURIComponent(url));
  const videoDetails = await videoDetailsResponse.json();
  const videoTitle = videoDetails.title;
  const duration = videoDetails.duration;

  videoDetailsDiv.style.display = 'flex';
  const newRow = document.createElement('tr');
  newRow.innerHTML = `
  <td>${videoTitle}</td>
  <td>${duration}</td>
`;
  videoDetailstable.appendChild(newRow);

  // Request the download and get the download token
  const tokenResponse = await fetch('/download?url=' + encodeURIComponent(url));
  if (tokenResponse.ok) {
    const tokenData = await tokenResponse.json();
    const downloadToken = tokenData.token;
    const health_check_period = tokenData.health_check_period;
    // Poll the backend for download completion
    const checkDownload = async () => {
      const downloadResponse = await fetch('/progress?token=' + encodeURIComponent(downloadToken));
      if (downloadResponse.status === 202) {
        setTimeout(checkDownload, health_check_period); // Poll every 5 seconds
      } else if (downloadResponse.status === 200) {
        // Download the file when the progress is 100%
        const fileResponse = await fetch('/downloadFile?token=' + encodeURIComponent(downloadToken));
        spinner.style.display = 'none';
        if (fileResponse.ok) {
          const blob = await fileResponse.blob();
          const downloadLink = document.createElement('a');
          downloadLink.href = URL.createObjectURL(blob);
          downloadLink.download = '/downloadFile?token=' + encodeURIComponent(downloadToken);

          // Show the download button
          const downloadButton = document.getElementById('download-btn');
          downloadButton.style.display = 'inline-block';
          downloadButton.onclick = () => {
            document.body.appendChild(downloadLink);
            downloadLink.click();
            document.body.removeChild(downloadLink);
            resultDiv.innerHTML = '<p>Download is completed</p>';
          };

          // Hide the spinner
          spinner.style.display = 'none';
          resultDiv.innerHTML = '<p>Download is ready:</p>';

        } else {
          const errorMessage = await fileResponse.text();
          resultDiv.innerHTML = '<p>Error: ' + errorMessage + '</p>';
        }
      }
    };
    resultDiv.innerHTML = '<p>Downloading ...</p>';
    checkDownload();
  } else {
    // Handle the error returned by the /download endpoint
    const errorMessage = await tokenResponse.text();
    spinner.style.display = 'none';
    resultDiv.innerHTML = '<p>Error: ' + errorMessage + '</p>';
  }

  // ... (rest of the JavaScript code)

});

// Read the URL query string
const queryString = window.location.search;
const urlParams = new URLSearchParams(queryString);

// Define an async function for form submission
async function submitForm() {
  const event = new Event('submit');
  await form.dispatchEvent(event);
}

// Check if the watch?v= parameter exists
if (urlParams.has('v')) {
  document.getElementById('url-input').value = 'https://www.youtube.com/watch?v=' + urlParams.get('v');
  submitForm();
}

