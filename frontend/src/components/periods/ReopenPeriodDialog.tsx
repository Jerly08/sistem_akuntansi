import React, { useState } from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  TextField,
  Typography,
  Box,
  Alert,
  CircularProgress,
} from '@mui/material';
import { Lock, LockOpen, Warning } from '@mui/icons-material';

export interface ReopenPeriodDialogProps {
  open: boolean;
  period: string; // Format: "2025-01"
  year: number;
  month: number;
  onClose: () => void;
  onReopen: (year: number, month: number, reason: string) => Promise<boolean>;
  isLoading?: boolean;
}

export const ReopenPeriodDialog: React.FC<ReopenPeriodDialogProps> = ({
  open,
  period,
  year,
  month,
  onClose,
  onReopen,
  isLoading = false,
}) => {
  const [reason, setReason] = useState('');
  const [error, setError] = useState('');

  const handleReopen = async () => {
    if (!reason.trim()) {
      setError('Alasan pembukaan kembali wajib diisi');
      return;
    }

    if (reason.trim().length < 10) {
      setError('Alasan minimal 10 karakter');
      return;
    }

    const success = await onReopen(year, month, reason.trim());
    
    if (success) {
      // Reset form
      setReason('');
      setError('');
    }
  };

  const handleClose = () => {
    if (!isLoading) {
      setReason('');
      setError('');
      onClose();
    }
  };

  const monthNames = [
    'Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni',
    'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember'
  ];

  return (
    <Dialog 
      open={open} 
      onClose={handleClose}
      maxWidth="sm"
      fullWidth
      disableEscapeKeyDown={isLoading}
    >
      <DialogTitle>
        <Box display="flex" alignItems="center" gap={1}>
          <LockOpen color="warning" />
          <span>Buka Kembali Periode</span>
        </Box>
      </DialogTitle>

      <DialogContent>
        <Box sx={{ mb: 3 }}>
          <Alert severity="warning" icon={<Warning />}>
            <Typography variant="body2">
              <strong>Perhatian:</strong> Anda akan membuka kembali periode yang sudah ditutup.
            </Typography>
          </Alert>
        </Box>

        <Box sx={{ mb: 3, p: 2, bgcolor: 'grey.100', borderRadius: 1 }}>
          <Typography variant="subtitle2" color="text.secondary" gutterBottom>
            Periode:
          </Typography>
          <Typography variant="h6" color="primary">
            {monthNames[month - 1]} {year}
          </Typography>
          <Typography variant="caption" color="text.secondary">
            ({period})
          </Typography>
        </Box>

        <Typography variant="body2" color="text.secondary" paragraph>
          Setelah dibuka kembali, transaksi dapat ditambahkan ke periode ini. 
          Pastikan untuk menutup kembali periode setelah selesai melakukan koreksi.
        </Typography>

        <TextField
          fullWidth
          multiline
          rows={4}
          label="Alasan Pembukaan Kembali *"
          placeholder="Contoh: Perlu menambahkan koreksi jurnal untuk invoice yang terlewat"
          value={reason}
          onChange={(e) => {
            setReason(e.target.value);
            setError('');
          }}
          error={!!error}
          helperText={error || 'Minimal 10 karakter. Alasan ini akan dicatat dalam audit log.'}
          disabled={isLoading}
          sx={{ mt: 2 }}
        />

        <Box sx={{ mt: 3, p: 2, bgcolor: 'info.light', borderRadius: 1, border: '1px solid', borderColor: 'info.main' }}>
          <Typography variant="caption" display="block" gutterBottom>
            <strong>Tips:</strong>
          </Typography>
          <Typography variant="caption" component="div">
            • Jelaskan secara spesifik alasan pembukaan kembali<br />
            • Sebutkan transaksi atau dokumen yang perlu ditambahkan<br />
            • Dokumentasi ini penting untuk audit trail
          </Typography>
        </Box>
      </DialogContent>

      <DialogActions sx={{ px: 3, pb: 2 }}>
        <Button 
          onClick={handleClose} 
          disabled={isLoading}
          color="inherit"
        >
          Batal
        </Button>
        <Button
          onClick={handleReopen}
          variant="contained"
          color="warning"
          disabled={isLoading || !reason.trim()}
          startIcon={isLoading ? <CircularProgress size={20} /> : <LockOpen />}
        >
          {isLoading ? 'Membuka...' : 'Buka Periode'}
        </Button>
      </DialogActions>
    </Dialog>
  );
};
