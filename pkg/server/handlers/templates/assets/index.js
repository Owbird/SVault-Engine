const cookBreadCrumbs = (path, container) => {
  const segments = path.split("/").filter(Boolean);

  const breadcrumbItems = segments.map((segment, index) => {
    const isLast = index === segments.length - 1;
    const url = "?dir=/" + segments.slice(0, index + 1).join("/");
    return isLast
      ? `<span class="text-gray-500">${segment}</span>`
      : `<a href="${url}" class="text-teal-600 hover:underline">${segment}</a>`;
  });

  container.innerHTML = `
          <ol class="flex space-x-2 text-sm">
            <li><a href="/" class="text-teal-600 hover:underline">Home</a></li>
            ${breadcrumbItems.map((item) => `<li>/ ${item}</li>`).join("")}
          </ol>
        `;
};

document.addEventListener("DOMContentLoaded", () => {
  const dropArea = document.getElementById("drop-area");
  const fileInput = document.getElementById("file-upload");
  const uploadButton = document.getElementById("upload-button");
  const uploadStatus = document.getElementById("upload-status");
  const breadcrumbsContainer = document.getElementById("breadcrumbs");

  const { searchParams } = new URL(window.location.href);

  const uploadDir = searchParams.get("dir") ?? "/";

  cookBreadCrumbs(uploadDir, breadcrumbsContainer);

  let files;

  const updateUploadBtnLabel = (total) => {
    if (total > 0) {
      uploadButton.innerText = `Upload ${total} file${total > 1 ? "s" : ""}`;
      uploadButton.classList.remove("hidden");
    }
  };

  const uploadFiles = async (files) => {
    const formData = new FormData();

    for (let file of files) {
      formData.append("file", file);
    }

    formData.append("uploadDir", uploadDir);

    return new Promise((resolve, reject) => {
      const xhr = new XMLHttpRequest();

      xhr.upload.addEventListener("progress", (event) => {
        if (event.lengthComputable) {
          const percentComplete = (event.loaded / event.total) * 100;
          uploadStatus.textContent = `Uploading: ${percentComplete.toFixed(2)}%`;
        }
      });

      xhr.addEventListener("load", () => {
        if (xhr.status === 200) {
          uploadStatus.textContent = "Upload complete!";
          window.location.reload();
          resolve();
        } else {
          uploadStatus.textContent = "Error uploading files.";
          reject(new Error("Upload failed"));
        }
      });

      xhr.addEventListener("error", () => {
        uploadStatus.textContent = "Network error during upload.";
        reject(new Error("Network error"));
      });

      xhr.open("POST", "/upload", true);
      xhr.send(formData);
    });
  };

  dropArea.addEventListener("dragover", (e) => {
    e.preventDefault();
    dropArea.classList.add("bg-teal-100");
  });

  dropArea.addEventListener("dragleave", () => {
    dropArea.classList.remove("bg-teal-100");
  });

  dropArea.addEventListener("drop", (e) => {
    e.preventDefault();
    dropArea.classList.remove("bg-teal-100");
    files = e.dataTransfer.files;
    updateUploadBtnLabel(files.length);
  });

  fileInput.addEventListener("change", () => {
    files = fileInput.files;
    updateUploadBtnLabel(files.length);
  });

  uploadButton.addEventListener("click", () => {
    if (files && files.length > 0) {
      uploadFiles(files);
    }
  });
});
