import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { Calendar, Clock, Users, ChevronRight, Filter } from 'lucide-react'
import * as Tabs from '@radix-ui/react-tabs'
import { api } from '../api/client'
import { Bill, CommitteeMeeting } from '../types'

export default function BillList() {
  const [bills, setBills] = useState<Bill[]>([])
  const [scheduledBills, setScheduledBills] = useState<Bill[]>([])
  const [meetings, setMeetings] = useState<CommitteeMeeting[]>([])
  const [loading, setLoading] = useState(true)
  const [filter, setFilter] = useState('')

  useEffect(() => {
    Promise.all([
      api.getBills(),
      api.getScheduledBills(),
      api.getMeetings(),
    ])
      .then(([allBills, scheduled, mtgs]) => {
        setBills(allBills)
        setScheduledBills(scheduled)
        setMeetings(mtgs)
        setLoading(false)
      })
      .catch((err) => {
        console.error('Failed to fetch data:', err)
        setLoading(false)
      })
  }, [])

  const filteredBills = bills.filter((bill) =>
    bill.title.toLowerCase().includes(filter.toLowerCase()) ||
    bill.number.toLowerCase().includes(filter.toLowerCase()) ||
    bill.sponsor.toLowerCase().includes(filter.toLowerCase())
  )

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    )
  }

  return (
    <div className="space-y-8">
      <div className="text-center space-y-2">
        <h1 className="text-3xl font-bold">Michigan Bill Tracker</h1>
        <p className="text-muted-foreground">
          Track bills scheduled for voting and contact your representatives
        </p>
      </div>

      <Tabs.Root defaultValue="scheduled" className="space-y-4">
        <Tabs.List className="flex gap-2 border-b">
          <Tabs.Trigger
            value="scheduled"
            className="px-4 py-2 text-sm font-medium border-b-2 border-transparent data-[state=active]:border-primary data-[state=active]:text-primary"
          >
            Scheduled Votes ({scheduledBills.length})
          </Tabs.Trigger>
          <Tabs.Trigger
            value="all"
            className="px-4 py-2 text-sm font-medium border-b-2 border-transparent data-[state=active]:border-primary data-[state=active]:text-primary"
          >
            All Bills ({bills.length})
          </Tabs.Trigger>
          <Tabs.Trigger
            value="meetings"
            className="px-4 py-2 text-sm font-medium border-b-2 border-transparent data-[state=active]:border-primary data-[state=active]:text-primary"
          >
            Committee Meetings ({meetings.length})
          </Tabs.Trigger>
        </Tabs.List>

        <Tabs.Content value="scheduled" className="space-y-4">
          {scheduledBills.length === 0 ? (
            <div className="text-center py-12 text-muted-foreground">
              No bills currently scheduled for voting
            </div>
          ) : (
            scheduledBills.map((bill) => (
              <BillCard key={bill.id} bill={bill} />
            ))
          )}
        </Tabs.Content>

        <Tabs.Content value="all" className="space-y-4">
          <div className="relative">
            <Filter className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <input
              type="text"
              placeholder="Filter bills by title, number, or sponsor..."
              value={filter}
              onChange={(e) => setFilter(e.target.value)}
              className="w-full pl-10 pr-4 py-2 border rounded-md bg-background"
            />
          </div>
          {filteredBills.map((bill) => (
            <BillCard key={bill.id} bill={bill} />
          ))}
        </Tabs.Content>

        <Tabs.Content value="meetings" className="space-y-4">
          {meetings.length === 0 ? (
            <div className="text-center py-12 text-muted-foreground">
              No upcoming committee meetings
            </div>
          ) : (
            meetings.map((meeting) => (
              <MeetingCard key={meeting.committee + meeting.date} meeting={meeting} />
            ))
          )}
        </Tabs.Content>
      </Tabs.Root>
    </div>
  )
}

function BillCard({ bill }: { bill: Bill }) {
  const voteDate = bill.scheduledVote
    ? new Date(bill.scheduledVote.date).toLocaleDateString('en-US', {
        month: 'short',
        day: 'numeric',
        year: 'numeric',
      })
    : null

  const isUrgent = bill.scheduledVote &&
    new Date(bill.scheduledVote.date).getTime() - Date.now() < 7 * 24 * 60 * 60 * 1000

  return (
    <div className={`border rounded-lg p-4 space-y-3 hover:shadow-md transition-shadow ${
      isUrgent ? 'border-orange-300 bg-orange-50' : 'bg-card'
    }`}>
      <div className="flex items-start justify-between gap-4">
        <div className="space-y-1 flex-1">
          <div className="flex items-center gap-2">
            <span className={`px-2 py-1 rounded text-xs font-medium ${
              bill.chamber === 'house' ? 'bg-blue-100 text-blue-800' : 'bg-green-100 text-green-800'
            }`}>
              {bill.chamber === 'house' ? 'House' : 'Senate'}
            </span>
            <span className="text-sm font-medium text-muted-foreground">{bill.number}</span>
            {isUrgent && (
              <span className="px-2 py-1 rounded text-xs font-medium bg-red-100 text-red-800">
                Vote Soon
              </span>
            )}
          </div>
          <h3 className="font-semibold leading-tight">{bill.title}</h3>
          <p className="text-sm text-muted-foreground line-clamp-2">{bill.description}</p>
        </div>
        <Link
          to={`/email/${bill.id}`}
          className="flex items-center gap-1 px-3 py-2 bg-primary text-primary-foreground rounded-md text-sm font-medium hover:bg-primary/90 shrink-0"
        >
          <Users className="h-4 w-4" />
          Contact Reps
        </Link>
      </div>

      <div className="flex items-center gap-4 text-sm text-muted-foreground">
        <span className="flex items-center gap-1">
          <Users className="h-3 w-3" />
          {bill.sponsor}
        </span>
        {voteDate && (
          <span className="flex items-center gap-1">
            <Calendar className="h-3 w-3" />
            Vote: {voteDate}
          </span>
        )}
        {bill.committee && (
          <span className="flex items-center gap-1">
            <Clock className="h-3 w-3" />
            {bill.committee}
          </span>
        )}
      </div>

      <Link
        to={`/bill/${bill.id}`}
        className="inline-flex items-center gap-1 text-sm text-primary hover:underline"
      >
        View Details
        <ChevronRight className="h-3 w-3" />
      </Link>
    </div>
  )
}

function MeetingCard({ meeting }: { meeting: CommitteeMeeting }) {
  const meetingDate = new Date(meeting.date).toLocaleDateString('en-US', {
    weekday: 'short',
    month: 'short',
    day: 'numeric',
  })

  return (
    <div className="border rounded-lg p-4 space-y-2 bg-card hover:shadow-md transition-shadow">
      <div className="flex items-center gap-2">
        <span className={`px-2 py-1 rounded text-xs font-medium ${
          meeting.chamber === 'house' ? 'bg-blue-100 text-blue-800' : 'bg-green-100 text-green-800'
        }`}>
          {meeting.chamber === 'house' ? 'House' : 'Senate'}
        </span>
        <h3 className="font-semibold">{meeting.committee}</h3>
      </div>
      <div className="flex items-center gap-4 text-sm text-muted-foreground">
        <span className="flex items-center gap-1">
          <Calendar className="h-3 w-3" />
          {meetingDate} at {meeting.time}
        </span>
        <span>{meeting.location}</span>
      </div>
      {meeting.bills.length > 0 && (
        <div className="text-sm">
          <span className="text-muted-foreground">Bills: </span>
          {meeting.bills.join(', ')}
        </div>
      )}
    </div>
  )
}
