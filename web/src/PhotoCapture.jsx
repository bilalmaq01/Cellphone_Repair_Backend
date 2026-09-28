import { useState } from 'react'

const MAX_PHOTO_EDGE = 1280
const JPEG_QUALITY = 0.72

function resizePhoto(file) {
  return new Promise((resolve, reject) => {
    const source = URL.createObjectURL(file)
    const image = new Image()

    image.onload = () => {
      URL.revokeObjectURL(source)
      const scale = Math.min(1, MAX_PHOTO_EDGE / Math.max(image.width, image.height))
      const canvas = document.createElement('canvas')
      canvas.width = Math.round(image.width * scale)
      canvas.height = Math.round(image.height * scale)
      const context = canvas.getContext('2d')
      if (!context) return reject(new Error('Could not process this photo.'))
      context.drawImage(image, 0, 0, canvas.width, canvas.height)
      resolve(canvas.toDataURL('image/jpeg', JPEG_QUALITY))
    }

    image.onerror = () => {
      URL.revokeObjectURL(source)
      reject(new Error('Could not read this photo.'))
    }
    image.src = source
  })
}

export default function PhotoCapture({ label, value, onChange }) {
  const [error, setError] = useState('')

  async function selectPhoto(event) {
    const file = event.target.files?.[0]
    if (!file) return
    if (!file.type.startsWith('image/')) {
      setError('Choose an image file.')
      onChange('')
      return
    }

    setError('')
    try {
      onChange(await resizePhoto(file))
    } catch (err) {
      onChange('')
      setError(err.message)
    }
  }

  return (
    <div className="photo-field">
      <label>{label} *</label>
      <input type="file" accept="image/*" capture="environment" onChange={selectPhoto} required />
      {error && <div className="error">{error}</div>}
      {value && <img className="photo-preview" src={value} alt={label + ' preview'} />}
    </div>
  )
}
