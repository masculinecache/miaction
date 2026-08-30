import { useState, useEffect } from 'react'
import { useParams, Link } from 'react-router-dom'
import { ArrowLeft, Send, MapPin, User, Mail, Eye } from 'lucide-react'
import * as Select from '@radix-ui/react-select'
import * as Checkbox from '@radix-ui/react-checkbox'
import { api } from '../api/client'
import { Bill, Representative, EmailTemplate } from '../types'

export default function EmailAction() {
  const { billId } = useParams<{ billId: string }>()
  const [bill, setBill] = useState<Bill | null>(null)
  const [templates, setTemplates] = useState<EmailTemplate[]>([])
  const [selectedTemplate, setSelectedTemplate] = useState('')
  const [variables, setVariables] = useState<Record<string, string>>({})
  const [zipCode, setZipCode] = useState('')
  const [reps, setReps] = useState<Representative[]>([])
  const [selectedReps, setSelectedReps] = useState<string[]>([])
  const [loading, setLoading] = useState(false)
  const [step, setStep] = useState<'location' | 'compose' | 'send'>('location')
  const [preview, setPreview] = useState<{ subject: string; body: string } | null>(null)
  const [mailtoLinks, setMailtoLinks] = useState<{ repId: string; name: string; mailto: string }[] | null>(null)

  useEffect(() => {
    if (!billId) return
    Promise.all([
      api.getBill(billId),
      api.getEmailTemplates(),
    ])
      .then(([billData, templateData]) => {
        setBill(billData)
        setTemplates(templateData)
        if (templateData.length > 0) {
          setSelectedTemplate(templateData[0].id)
          initVariables(templateData[0], billData)
        }
      })
      .catch((err) => console.error('Failed to fetch data:', err))
  }, [billId])

  const initVariables = (template: EmailTemplate, billData: Bill) => {
    const vars: Record<string, string> = {}
    template.variables.forEach((v) => {
      if (v.name === 'BillNumber') vars[v.name] = billData.number
      else if (v.name === 'BillTitle') vars[v.name] = billData.title
      else if (v.default) vars[v.name] = v.default
    })
    setVariables(vars)
  }

  const handleTemplateChange = (templateId: string) => {
    setSelectedTemplate(templateId)
    const template = templates.find((t) => t.id === templateId)
    if (template && bill) {
      initVariables(template, bill)
    }
  }

  const findReps = async () => {
    if (!zipCode) return
    setLoading(true)
    try {
      const data = await api.findRepresentatives({
        street: '',
        city: '',
        state: 'MI',
        zip: zipCode,
      })
      setReps(data)
      setSelectedReps(data.map((r) => r.id))
      setStep('compose')
    } catch (err) {
      console.error('Failed to find representatives:', err)
      alert('Failed to find representatives. Please check your ZIP code.')
    } finally {
      setLoading(false)
    }
  }

  const handlePreview = async () => {
    if (!selectedTemplate) return
    try {
      const data = await api.previewEmail(selectedTemplate, variables)
      setPreview(data)
    } catch (err) {
      console.error('Failed to preview email:', err)
    }
  }

  const handleSend = async () => {
    if (!selectedTemplate || selectedReps.length === 0) return
    setLoading(true)
    try {
      const data = await api.composeEmail(selectedTemplate, variables, selectedReps)
      setMailtoLinks(data)
      setStep('send')
    } catch (err) {
      console.error('Failed to compose email:', err)
      alert('Failed to compose email. Please check all required fields.')
    } finally {
      setLoading(false)
    }
  }

  const toggleRep = (repId: string) => {
    setSelectedReps((prev) =>
      prev.includes(repId) ? prev.filter((id) => id !== repId) : [...prev, repId]
    )
  }

  const currentTemplate = templates.find((t) => t.id === selectedTemplate)

  return (
    <div className="space-y-6">
      <Link
        to="/"
        className="inline-flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground"
      >
        <ArrowLeft className="h-4 w-4" />
        Back to Bills
      </Link>

      {bill && (
        <div className="border rounded-lg p-4 bg-card">
          <span className="text-sm font-medium text-muted-foreground">{bill.number}</span>
          <h1 className="text-lg font-semibold">{bill.title}</h1>
        </div>
      )}

      {step === 'location' && (
        <div className="border rounded-lg p-6 space-y-4 bg-card max-w-md">
          <div className="space-y-2">
            <h2 className="text-xl font-semibold">Find Your Representatives</h2>
            <p className="text-sm text-muted-foreground">
              Enter your ZIP code to find your Michigan House and Senate representatives
            </p>
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium flex items-center gap-2">
              <MapPin className="h-4 w-4" />
              ZIP Code
            </label>
            <input
              type="text"
              value={zipCode}
              onChange={(e) => setZipCode(e.target.value)}
              placeholder="e.g., 48226"
              className="w-full px-3 py-2 border rounded-md bg-background"
              maxLength={10}
            />
          </div>
          <button
            onClick={findReps}
            disabled={!zipCode || loading}
            className="w-full px-4 py-2 bg-primary text-primary-foreground rounded-md font-medium hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {loading ? 'Finding...' : 'Find Representatives'}
          </button>
        </div>
      )}

      {step === 'compose' && (
        <div className="space-y-6">
          <div className="border rounded-lg p-4 space-y-4 bg-card">
            <h2 className="text-xl font-semibold">Your Representatives</h2>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {reps.map((rep) => (
                <div
                  key={rep.id}
                  className={`border rounded-lg p-4 space-y-2 cursor-pointer transition-colors ${
                    selectedReps.includes(rep.id) ? 'border-primary bg-primary/5' : 'hover:bg-accent'
                  }`}
                  onClick={() => toggleRep(rep.id)}
                >
                  <div className="flex items-center gap-3">
                    <Checkbox.Root
                      checked={selectedReps.includes(rep.id)}
                      onCheckedChange={() => toggleRep(rep.id)}
                      className="w-4 h-4 border rounded flex items-center justify-center data-[state=checked]:bg-primary data-[state=checked]:border-primary"
                    >
                      <Checkbox.Indicator className="text-primary-foreground">
                        <svg width="10" height="10" viewBox="0 0 10 10" fill="none">
                          <path d="M1 5L4 8L9 2" stroke="currentColor" strokeWidth="2" />
                        </svg>
                      </Checkbox.Indicator>
                    </Checkbox.Root>
                    <div>
                      <div className="font-medium">{rep.name}</div>
                      <div className="text-sm text-muted-foreground">
                        {rep.chamber === 'house' ? 'House' : 'Senate'} District {rep.district} ({rep.party})
                      </div>
                    </div>
                  </div>
                  <div className="text-sm space-y-1 pl-7">
                    <div className="flex items-center gap-2">
                      <Mail className="h-3 w-3 text-muted-foreground" />
                      {rep.email}
                    </div>
                    <div className="flex items-center gap-2">
                      <MapPin className="h-3 w-3 text-muted-foreground" />
                      {rep.office}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>

          <div className="border rounded-lg p-4 space-y-4 bg-card">
            <h2 className="text-xl font-semibold">Compose Email</h2>
            
            <div className="space-y-2">
              <label className="text-sm font-medium">Email Template</label>
              <Select.Root value={selectedTemplate} onValueChange={handleTemplateChange}>
                <Select.Trigger className="w-full px-3 py-2 border rounded-md bg-background flex items-center justify-between">
                  <Select.Value />
                  <span className="text-muted-foreground">▼</span>
                </Select.Trigger>
                <Select.Portal>
                  <Select.Content className="bg-popover border rounded-md shadow-md">
                    <Select.Viewport className="p-1">
                      {templates.map((template) => (
                        <Select.Item
                          key={template.id}
                          value={template.id}
                          className="px-3 py-2 rounded-md cursor-pointer hover:bg-accent outline-none"
                        >
                          <Select.ItemText>{template.name}</Select.ItemText>
                        </Select.Item>
                      ))}
                    </Select.Viewport>
                  </Select.Content>
                </Select.Portal>
              </Select.Root>
              {currentTemplate && (
                <p className="text-sm text-muted-foreground">{currentTemplate.description}</p>
              )}
            </div>

            {currentTemplate?.variables.map((variable) => (
              <div key={variable.name} className="space-y-2">
                <label className="text-sm font-medium flex items-center gap-1">
                  {variable.name}
                  {variable.required && <span className="text-red-500">*</span>}
                </label>
                {variable.name === 'PersonalMessage' ? (
                  <textarea
                    value={variables[variable.name] || ''}
                    onChange={(e) =>
                      setVariables((prev) => ({ ...prev, [variable.name]: e.target.value }))
                    }
                    placeholder={variable.description}
                    rows={4}
                    className="w-full px-3 py-2 border rounded-md bg-background resize-y"
                  />
                ) : (
                  <input
                    type="text"
                    value={variables[variable.name] || ''}
                    onChange={(e) =>
                      setVariables((prev) => ({ ...prev, [variable.name]: e.target.value }))
                    }
                    placeholder={variable.description}
                    className="w-full px-3 py-2 border rounded-md bg-background"
                  />
                )}
              </div>
            ))}

            <div className="flex gap-4 pt-4">
              <button
                onClick={handlePreview}
                className="flex items-center gap-2 px-4 py-2 border rounded-md font-medium hover:bg-accent"
              >
                <Eye className="h-4 w-4" />
                Preview
              </button>
              <button
                onClick={handleSend}
                disabled={selectedReps.length === 0 || loading}
                className="flex items-center gap-2 px-4 py-2 bg-primary text-primary-foreground rounded-md font-medium hover:bg-primary/90 disabled:opacity-50"
              >
                <Send className="h-4 w-4" />
                {loading ? 'Composing...' : 'Generate Emails'}
              </button>
            </div>

            {preview && (
              <div className="border rounded-lg p-4 space-y-2 bg-muted">
                <h3 className="font-medium">Preview</h3>
                <div className="text-sm">
                  <span className="font-medium">Subject:</span> {preview.subject}
                </div>
                <pre className="text-sm whitespace-pre-wrap bg-background p-3 rounded border">
                  {preview.body}
                </pre>
              </div>
            )}
          </div>
        </div>
      )}

      {step === 'send' && mailtoLinks && (
        <div className="border rounded-lg p-6 space-y-4 bg-card">
          <h2 className="text-xl font-semibold">Send Your Emails</h2>
          <p className="text-muted-foreground">
            Click each button below to open your email client with a pre-composed message to each representative.
          </p>
          <div className="space-y-3">
            {mailtoLinks.map((link) => (
              <a
                key={link.repId}
                href={link.mailto}
                className="flex items-center justify-between p-4 border rounded-lg hover:bg-accent transition-colors"
              >
                <div className="flex items-center gap-3">
                  <User className="h-5 w-5 text-muted-foreground" />
                  <span className="font-medium">{link.name}</span>
                </div>
                <span className="flex items-center gap-2 text-sm text-primary">
                  <Mail className="h-4 w-4" />
                  Open Email
                </span>
              </a>
            ))}
          </div>
          <div className="flex gap-4 pt-4">
            <button
              onClick={() => setStep('compose')}
              className="px-4 py-2 border rounded-md font-medium hover:bg-accent"
            >
              Back to Compose
            </button>
            <Link
              to="/"
              className="px-4 py-2 bg-primary text-primary-foreground rounded-md font-medium hover:bg-primary/90"
            >
              Back to Bills
            </Link>
          </div>
        </div>
      )}
    </div>
  )
}
