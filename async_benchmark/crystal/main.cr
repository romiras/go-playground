# Get the number of fibers from command-line arguments, defaulting to 100,000
num_fibers = ARGV.size > 0 ? ARGV[0].to_i : 100_000

# Create and start fibers
num_fibers.times do
  spawn do
    sleep 10.seconds
  end
end

# Wait for all fibers to finish
Fiber.yield

puts "All fibers finished."
