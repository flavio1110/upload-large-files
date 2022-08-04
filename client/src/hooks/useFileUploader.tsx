import { useState } from "react";
import { FileToUpload } from "../App";

const useFileUploader = () => {
  const [files, setFiles] = useState<FileToUpload[]>([]);

  const onDropFiles = (newFiles: File[]) => {
    const newFilesToUpload = newFiles
      .filter((f) => !files.some((ef) => ef.file.name === f.name))
      .map((f) => {
        return { file: f, progress: 0, status: "waiting" as const };
      });

    setFiles([...files, ...newFilesToUpload]);
  };

  return { files, onDropFiles };
};

export default useFileUploader;
