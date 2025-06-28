# frozen_string_literal: true

# Default number of fibers
num_routines = 100_000

# Parse command line arguments
if ARGV.length > 0
  num_routines = ARGV[0].to_i
end

fibers = []

# Create fibers
num_routines.times do
  fibers << Fiber.new do
    sleep(10)
  end
end

# Run all fibers
fibers.each(&:resume)

# Wait for all fibers to finish (using a simple loop to simulate waiting)
fibers.each do |fiber|
  while fiber.alive?
    Fiber.yield # Yield control back to the scheduler
  end
end

puts "All fibers finished."
