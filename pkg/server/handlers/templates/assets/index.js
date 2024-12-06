document.addEventListener("DOMContentLoaded", () => {
  const dropArea = document.getElementById("drop-area");
  const fileInput = document.getElementById("file-upload");
  const uploadButton = document.getElementById("upload-button");
  const uploadStatus = document.getElementById("upload-status");

  const { searchParams } = new URL(window.location.href);

  const uploadDir = searchParams.get("dir") ?? "/";

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

    uploadStatus.textContent = "Uploading...";

    try {
      await fetch("/upload", {
        mode: "no-cors",
        method: "POST",
        body: formData,
      });
      uploadStatus.textContent = "Upload complete!";
      window.location.reload();
    } catch (error) {
      uploadStatus.textContent = "Error uploading files.";
    }
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
