import React from 'react';

interface SuccessAlertProps {
  message: string;
  onClose?: () => void;
}

export const SuccessAlert: React.FC<SuccessAlertProps> = ({
  message,
  onClose,
}) => {
  return (
    <div
      className="alert alert-success alert-dismissible fade show"
      role="alert"
    >
      <strong>Success:</strong> {message}
      {onClose && (
        <button
          type="button"
          className="btn-close"
          onClick={onClose}
          aria-label="Close"
        ></button>
      )}
    </div>
  );
};

