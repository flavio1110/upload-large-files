import Dropzone from "./components/Dropzone";
import SelectedFiles from "./components/SelectedFiles";
import useFileUploader from "./hooks/useFileUploader";

export interface FileToUpload {
  file: File;
  status: "waiting" | "uploading" | "uploaded" | "failed";
  progress: number;
  error?: string;
  id?: string;
}

function App() {
  const { files, onDropFiles } = useFileUploader();
  return (
    <div>
      <div>
        <Dropzone onDropFiles={onDropFiles} />
        <SelectedFiles files={files} />
      </div>
    </div>
  );
}

export default App;
