import { useCallback } from "react";
import { useDropzone } from "react-dropzone";

interface Props {
  onDropFiles: (file: File[]) => void;
}

const Dropzone = ({ onDropFiles }: Props) => {
  const onDrop = useCallback(
    (acceptedFiles: File[]) => {
      onDropFiles(acceptedFiles);
    },
    [onDropFiles]
  );
  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
  });

  return (
    <div
      {...getRootProps()}
      style={{ margin: "20px", border: "solid 1px black" }}
    >
      <input {...getInputProps()} />
      {isDragActive ? (
        <p>Drop the files here ...</p>
      ) : (
        <p>Drag 'n' drop some files here, or click to select files</p>
      )}
    </div>
  );
};
export default Dropzone;
