import { useEffect, useState } from "react";
import api from "../api";
import { FileToUpload } from "../App";

interface Props {
  files: FileToUpload[];
}
const SelectedFiles = ({ files }: Props) => {
  return (
    <>
      <h3>Files ({files.length})</h3>
      {files.length === 0 && <p>Select some files to upload</p>}
      {files.length > 0 &&
        files.map((f, i) => <SelectedFile key={i} selected={f} />)}
    </>
  );
};

interface SelectedFileProps {
  selected: FileToUpload;
}

const SelectedFile = ({ selected }: SelectedFileProps) => {
  const [sel, setSel] = useState(selected);
  useEffect(() => {
    const prepare = async () => {
      const res = await api.Prepare(sel.file.name, sel.file.type);

      if (res?.status === 200) {
        setSel({ ...sel, status: "uploading", progress: 25, id: res?.data.id });
      } else {
        setSel({ ...sel, status: "failed", error: "fail to prepare" });
      }
    };
    prepare();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <>
      <h4>
        {sel.id ? sel.id : "Not prepared"} - {sel.file.name}
      </h4>
      <p>
        Status: {sel.status} - Progress {sel.progress}%
      </p>
      {sel.error && <span style={{ color: "red" }}>{sel.error}</span>}
      <hr />
    </>
  );
};

export default SelectedFiles;
