import { useState, useEffect } from 'react'
import { useParams, Link } from 'react-router-dom'
import { Calendar, Users, Clock, ArrowLeft, Mail, FileText, ExternalLink, BookText } from 'lucide-react'
import { api } from '../api/client'
import { Bill } from '../types'

export default function BillDetail() {
  const { id } = useParams<{ id: string }>()
  const [bill, setBill] = useState<Bill | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (!id) return
    api.getBill(id)
      .then((data) => {
        setBill(data)
        setLoading(false)
      })
      .catch((err) => {
        console.error('Failed to fetch bill:', err)
        setLoading(false)
      })
  }, [id])

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    )
  }

  if (!bill) {
    return (
      <div className="text-center py-12">
        <p className="text-muted-foreground">Bill not found</p>
        <Link to="/" className="text-primary hover:underline mt-4 inline-block">
          Back to bills
        </Link>
      </div>
    )
  }

  const voteDate = bill.scheduledVote
    ? new Date(bill.scheduledVote.date).toLocaleDateString('en-US', {
        weekday: 'long',
        month: 'long',
        day: 'numeric',
        year: 'numeric',
      })
    : null

  return (
    <div className="space-y-6">
      <Link
        to="/"
        className="inline-flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground"
      >
        <ArrowLeft className="h-4 w-4" />
        Back to Bills
      </Link>

      <div className="border rounded-lg p-6 space-y-4 bg-card">
        <div className="flex items-center gap-2">
          <span className={`px-2 py-1 rounded text-xs font-medium ${
            bill.chamber === 'house' ? 'bg-blue-100 text-blue-800' : 'bg-green-100 text-green-800'
          }`}>
            {bill.chamber === 'house' ? 'House' : 'Senate'}
          </span>
          <span className="text-sm font-medium text-muted-foreground">{bill.number}</span>
        </div>

        <h1 className="text-2xl font-bold">{bill.title}</h1>
        <p className="text-muted-foreground">{bill.description}</p>
        {bill.subject && (
          <p className="text-sm text-muted-foreground bg-muted/50 border-l-2 border-primary/40 pl-3 py-2 rounded-sm">
            {bill.subject}
          </p>
        )}

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 pt-4 border-t">
          <div className="space-y-2">
            <div className="flex items-center gap-2 text-sm">
              <Users className="h-4 w-4 text-muted-foreground" />
              <span className="font-medium">Sponsor:</span>
              <span>{bill.sponsor}</span>
            </div>
            {bill.coSponsors.length > 0 && (
              <div className="flex items-center gap-2 text-sm">
                <Users className="h-4 w-4 text-muted-foreground" />
                <span className="font-medium">Co-sponsors:</span>
                <span>{bill.coSponsors.join(', ')}</span>
              </div>
            )}
            {bill.committee && (
              <div className="flex items-center gap-2 text-sm">
                <Clock className="h-4 w-4 text-muted-foreground" />
                <span className="font-medium">Committee:</span>
                <span>{bill.committee}</span>
              </div>
            )}
          </div>

          <div className="space-y-2">
            <div className="flex items-center gap-2 text-sm">
              <Calendar className="h-4 w-4 text-muted-foreground" />
              <span className="font-medium">Introduced:</span>
              <span>{new Date(bill.introducedDate).toLocaleDateString()}</span>
            </div>
            <div className="flex items-center gap-2 text-sm">
              <Clock className="h-4 w-4 text-muted-foreground" />
              <span className="font-medium">Last Action:</span>
              <span>{bill.lastAction}</span>
            </div>
            {voteDate && (
              <div className="flex items-center gap-2 text-sm bg-orange-50 p-2 rounded">
                <Calendar className="h-4 w-4 text-orange-600" />
                <span className="font-medium text-orange-800">Scheduled Vote:</span>
                <span className="text-orange-800">{voteDate}</span>
              </div>
            )}
          </div>
        </div>

        <div className="flex gap-4 pt-4">
          <Link
            to={`/email/${bill.id}`}
            className="inline-flex items-center gap-2 px-4 py-2 bg-primary text-primary-foreground rounded-md font-medium hover:bg-primary/90"
          >
            <Mail className="h-4 w-4" />
            Contact Your Representatives
          </Link>
          <a
            href={bill.url}
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex items-center gap-2 px-4 py-2 border rounded-md font-medium hover:bg-accent"
          >
            View on Legislature Site
          </a>
        </div>
      </div>

      {bill.analysisDocuments && bill.analysisDocuments.length > 0 && (
        <div className="border rounded-lg p-6 space-y-4 bg-card">
          <h2 className="text-lg font-semibold flex items-center gap-2">
            <FileText className="h-5 w-5 text-muted-foreground" />
            House Fiscal Agency Analysis
          </h2>
          <div className="space-y-3">
            {bill.analysisDocuments.map((doc, idx) => (
              <div key={idx} className="border-l-2 border-primary/30 pl-4 py-1">
                <p className="text-sm font-medium">{doc.title}</p>
                {doc.description && (
                  <p className="text-sm text-muted-foreground mt-0.5">{doc.description}</p>
                )}
              </div>
            ))}
          </div>
        </div>
      )}

      {bill.billDocuments && bill.billDocuments.length > 0 && (
        <div className="border rounded-lg p-6 space-y-4 bg-card">
          <h2 className="text-lg font-semibold flex items-center gap-2">
            <BookText className="h-5 w-5 text-muted-foreground" />
            Bill Text
          </h2>
          <div className="space-y-3">
            {bill.billDocuments.map((doc, idx) => (
              <div key={idx} className="border-l-2 border-primary/30 pl-4 py-1">
                <p className="text-sm font-medium">{doc.title}</p>
                {doc.description && (
                  <p className="text-xs text-muted-foreground mt-0.5 mb-2">{doc.description}</p>
                )}
                <div className="flex gap-3 mt-1">
                  {doc.htmlUrl && (
                    <a
                      href={doc.htmlUrl}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="inline-flex items-center gap-1 text-xs text-primary hover:underline"
                    >
                      <ExternalLink className="h-3 w-3" />
                      View HTML
                    </a>
                  )}
                  {doc.pdfUrl && (
                    <a
                      href={doc.pdfUrl}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="inline-flex items-center gap-1 text-xs text-primary hover:underline"
                    >
                      <FileText className="h-3 w-3" />
                      View PDF
                    </a>
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
